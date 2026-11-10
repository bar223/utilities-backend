package handler

import (
	"errors"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"utilities-backend/internal/app/ds"
	"utilities-backend/internal/app/repository"
)

// maxUploadSize — лимит на размер данных формы с файлами (картинка +
// короткое видео вместе), читаемых в память при разборе multipart-формы.
const maxUploadSize = 20 << 20 // 20 МБ

// parseUtilityID читает id услуги из пути. При ошибке сам отвечает 400.
func (h *Handler) parseUtilityID(ctx *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("некорректный идентификатор услуги"))
		return 0, false
	}
	return uint(id), true
}

// GetUtilities godoc
// @Summary      Список коммунальных услуг
// @Description  Опубликованные услуги с тарифом не выше tariff. Доступно гостю; с токеном isMine = 1 у услуг текущего пользователя.
// @Tags         utilities
// @Produce      json
// @Param        tariff  query     number  false  "тариф не выше, руб."
// @Success      200     {array}   ds.UtilityListItem
// @Failure      400     {object}  errorResponse
// @Router       /utilities [get]
func (h *Handler) GetUtilities(ctx *gin.Context) {
	var maxTariffRate *float64
	if tariffQuery := ctx.Query("tariff"); tariffQuery != "" {
		parsed, err := strconv.ParseFloat(tariffQuery, 64)
		if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
			h.errorHandler(ctx, http.StatusBadRequest, errors.New("параметр tariff должен быть числом"))
			return
		}
		maxTariffRate = &parsed
	}

	utilities, err := h.Repository.GetUtilities(maxTariffRate)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	items := make([]ds.UtilityListItem, 0, len(utilities))
	for _, u := range utilities {
		likes, err := h.Repository.GetLikesCount(u.ID)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
		items = append(items, ds.UtilityListItem{
			Utility:    u,
			IsMine:     ds.BoolToInt(u.CreatorID == currentUserID(ctx)),
			LikesCount: likes,
		})
	}

	ctx.JSON(http.StatusOK, items)
}

// GetUtilityFeed godoc
// @Summary      Лента коммунальных услуг (без id)
// @Description  Первая опубликованная услуга с вложенным создателем — вход в ленту из панели вкладок, id не указывается. Доступно гостю; с токеном liked = 1, если пользователь лайкнул услугу.
// @Tags         utilities
// @Produce      json
// @Success      200   {object}  ds.UtilityFeedItem
// @Failure      404   {object}  errorResponse
// @Router       /utilities/feed [get]
func (h *Handler) GetUtilityFeed(ctx *gin.Context) {
	h.utilityFeed(ctx)
}

// GetUtilityFeedByID godoc
// @Summary      Лента коммунальных услуг по id
// @Description  Опубликованная услуга по id с вложенным создателем, с next=true — следующая после неё (после последней — снова первая). Доступно гостю; с токеном liked = 1, если пользователь лайкнул услугу.
// @Tags         utilities
// @Produce      json
// @Param        id    path      int     true   "id услуги"
// @Param        next  query     bool    false  "следующая после id"
// @Success      200   {object}  ds.UtilityFeedItem
// @Failure      400   {object}  errorResponse
// @Failure      404   {object}  errorResponse
// @Router       /utilities/feed/{id} [get]
func (h *Handler) GetUtilityFeedByID(ctx *gin.Context) {
	h.utilityFeed(ctx)
}

// utilityFeed — общий метод ленты для обоих адресов: id берётся из пути,
// без id — первая опубликованная услуга.
func (h *Handler) utilityFeed(ctx *gin.Context) {
	var id uint
	if ctx.Param("id") != "" {
		parsed, ok := h.parseUtilityID(ctx)
		if !ok {
			return
		}
		id = parsed
	}
	next := ctx.Query("next") == "true"

	utility, err := h.Repository.GetUtilityFeedItem(id, next)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, errors.New("услуга не найдена"))
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	liked, err := h.Repository.IsLikedByUser(currentUserID(ctx), utility.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	likes, err := h.Repository.GetLikesCount(utility.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.UtilityFeedItem{
		Utility:    utility,
		Creator:    ds.UserSerializer{ID: utility.Creator.ID, Login: utility.Creator.Login},
		Liked:      ds.BoolToInt(liked),
		LikesCount: likes,
	})
}

// GetUtilityDraft godoc
// @Summary      Черновик текущего пользователя
// @Tags         utilities
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  ds.Utility
// @Failure      401  {object}  errorResponse
// @Failure      404  {object}  errorResponse  "черновика нет"
// @Router       /utilities/draft [get]
func (h *Handler) GetUtilityDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetUtilityDraft(currentUserID(ctx))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, errors.New("черновика нет"))
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, draft)
}

// CreateUtility godoc
// @Summary      Добавление услуги с изображением и видео
// @Description  Создаёт черновик; создатель берётся из JWT. Файлы сохраняются в MinIO под латинскими именами.
// @Tags         utilities
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        name   formData  string  true   "название"
// @Param        image  formData  file    false  "изображение"
// @Param        video  formData  file    false  "короткое видео"
// @Success      201    {object}  ds.Utility
// @Failure      400    {object}  errorResponse
// @Failure      401    {object}  errorResponse
// @Failure      409    {object}  errorResponse  "черновик уже есть"
// @Router       /utilities [post]
func (h *Handler) CreateUtility(ctx *gin.Context) {
	if err := ctx.Request.ParseMultipartForm(maxUploadSize); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	name := ctx.Request.FormValue("name")
	if name == "" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("поле name обязательно"))
		return
	}

	// Файлы проверяются до создания черновика, чтобы неверный файл не
	// оставил после себя пустой черновик.
	files := map[string]*multipart.FileHeader{}
	for _, field := range []string{"image", "video"} {
		header, err := ctx.FormFile(field)
		if errors.Is(err, http.ErrMissingFile) {
			continue
		}
		if err == nil {
			_, err = repository.DetectFileType(field, header)
		}
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		files[field] = header
	}

	draft, err := h.Repository.CreateUtilityDraft(currentUserID(ctx), name)
	if err != nil {
		if errors.Is(err, repository.ErrDraftExists) {
			h.errorHandler(ctx, http.StatusConflict, err)
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if header, ok := files["image"]; ok {
		if draft.ImageURL, err = h.Repository.UploadUtilityImage(draft.ID, header); err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}
	if header, ok := files["video"]; ok {
		if draft.VideoURL, err = h.Repository.UploadUtilityVideo(draft.ID, header); err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	ctx.JSON(http.StatusCreated, draft)
}

// PublishUtility godoc
// @Summary      Публикация услуги
// @Description  Заполняет поля черновика текущего пользователя, статус draft → published.
// @Tags         utilities
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int                       true  "id черновика"
// @Param        body  body      ds.PublishUtilityRequest  true  "поля услуги"
// @Success      200   {object}  ds.Utility
// @Failure      400   {object}  errorResponse
// @Failure      401   {object}  errorResponse
// @Failure      404   {object}  errorResponse
// @Router       /utilities/{id} [put]
func (h *Handler) PublishUtility(ctx *gin.Context) {
	id, ok := h.parseUtilityID(ctx)
	if !ok {
		return
	}

	var req ds.PublishUtilityRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	utility, err := h.Repository.PublishUtility(id, currentUserID(ctx), req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, errors.New("черновик не найден"))
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, utility)
}

// DeleteUtility godoc
// @Summary      Удаление услуги (soft delete)
// @Tags         utilities
// @Security     BearerAuth
// @Param        id   path  int  true  "id услуги"
// @Success      204
// @Failure      401  {object}  errorResponse
// @Failure      404  {object}  errorResponse
// @Router       /utilities/{id} [delete]
func (h *Handler) DeleteUtility(ctx *gin.Context) {
	id, ok := h.parseUtilityID(ctx)
	if !ok {
		return
	}

	if err := h.Repository.DeleteUtility(id, currentUserID(ctx)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, errors.New("услуга не найдена или принадлежит другому пользователю"))
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// SetUtilityLike godoc
// @Summary      Лайк услуги
// @Description  value: 1 — поставить, 0 — снять лайк текущего пользователя.
// @Tags         utilities
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path      int             true  "id услуги"
// @Param        body  body      ds.LikeRequest  true  "0 или 1"
// @Success      200   {object}  ds.UtilityLikeResponse
// @Failure      400   {object}  errorResponse
// @Failure      401   {object}  errorResponse
// @Failure      404   {object}  errorResponse
// @Router       /utilities/{id}/like [post]
func (h *Handler) SetUtilityLike(ctx *gin.Context) {
	id, ok := h.parseUtilityID(ctx)
	if !ok {
		return
	}

	var req ds.LikeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("value должно быть 0 или 1"))
		return
	}

	published, err := h.Repository.IsUtilityPublished(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if !published {
		h.errorHandler(ctx, http.StatusNotFound, errors.New("услуга не найдена"))
		return
	}

	if err := h.Repository.SetLike(currentUserID(ctx), id, *req.Value); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	likes, err := h.Repository.GetLikesCount(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.UtilityLikeResponse{Liked: *req.Value, LikesCount: likes})
}

package handler

import (
	"errors"
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

// GetUtilities — GET /api/utilities?tariff=: список опубликованных услуг с
// фильтром по тарифу (не выше указанного). isMine = 1, если создатель
// услуги совпадает с текущим пользователем.
func (h *Handler) GetUtilities(ctx *gin.Context) {
	var maxTariffRate *float64
	if tariffQuery := ctx.Query("tariff"); tariffQuery != "" {
		parsed, err := strconv.ParseFloat(tariffQuery, 64)
		if err != nil {
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
			IsMine:     ds.BoolToInt(u.CreatorID == CurrentUserID()),
			LikesCount: likes,
		})
	}

	ctx.JSON(http.StatusOK, items)
}

// GetUtilityFeed — GET /api/utilities/feed?id=&next=: одна опубликованная
// услуга с вложенным создателем. Без id — первая услуга (вход в ленту из
// панели вкладок, id не указывается по методичке), с id — эта услуга,
// с id и next=true — следующая за ней. liked = 1, если текущий
// пользователь лайкнул услугу.
func (h *Handler) GetUtilityFeed(ctx *gin.Context) {
	var id uint
	if idQuery := ctx.Query("id"); idQuery != "" {
		parsed, err := strconv.ParseUint(idQuery, 10, 64)
		if err != nil || parsed == 0 {
			h.errorHandler(ctx, http.StatusBadRequest, errors.New("некорректный идентификатор услуги"))
			return
		}
		id = uint(parsed)
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

	liked, err := h.Repository.IsLikedByUser(CurrentUserID(), utility.ID)
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

// GetUtilityDraft — GET /api/utilities/draft: черновик текущего
// пользователя (не более одного), id не указывается. Нет черновика — 404.
func (h *Handler) GetUtilityDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetUtilityDraft(CurrentUserID())
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

// CreateUtility — POST /api/utilities (multipart/form-data): создаёт
// черновик по названию (name) и загружает файлы image и video в MinIO.
// Имена файлов генерируются на латинице, ссылки сохраняются в image_url и
// video_url. Второй черновик создать нельзя — 409.
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

	draft, err := h.Repository.CreateUtilityDraft(CurrentUserID(), name)
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

// PublishUtility — PUT /api/utilities/:id: заполняет поля черновика
// текущего пользователя и меняет статус на «опубликован». Вернуть в
// черновик нельзя; не черновик или чужая услуга — 404.
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

	utility, err := h.Repository.PublishUtility(id, CurrentUserID(), req)
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

// DeleteUtility — DELETE /api/utilities/:id: логическое удаление (soft
// delete) только услуги текущего пользователя. Успех — 204 без тела.
func (h *Handler) DeleteUtility(ctx *gin.Context) {
	id, ok := h.parseUtilityID(ctx)
	if !ok {
		return
	}

	if err := h.Repository.DeleteUtility(id, CurrentUserID()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, errors.New("услуга не найдена или принадлежит другому пользователю"))
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// SetUtilityLike — POST /api/utilities/:id/like: JSON {"value": 0|1} —
// снять или поставить лайк текущего пользователя на опубликованной услуге.
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

	if err := h.Repository.SetLike(CurrentUserID(), id, *req.Value); err != nil {
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

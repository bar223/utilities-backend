package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"utility-bills-backend/internal/app/ds"
)

// maxUploadSize — лимит на размер данных формы с файлами (картинка +
// короткое видео вместе), читаемых в память при разборе multipart-формы.
const maxUploadSize = 20 << 20 // 20 МБ

// APIListUtilities — GET /api/utilities: список опубликованных услуг с
// фильтром по тарифу (?tariff=) и признаком "это моя услуга" для
// текущего пользователя.
func (h *Handler) APIListUtilities(ctx *gin.Context) {
	tariffQuery := ctx.Query("tariff")

	var maxTariffRate *float64
	if tariffQuery != "" {
		if parsed, err := strconv.ParseFloat(tariffQuery, 64); err == nil {
			maxTariffRate = &parsed
		}
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
			IsMine:     u.CreatorID == CurrentUserID(),
			LikesCount: likes,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": items})
}

// APIFeed — GET /api/utilities/feed и /api/utilities/feed/:id: одна
// опубликованная услуга, ?next=true — следующая. Признак "liked" — лайкнул
// ли текущий пользователь.
func (h *Handler) APIFeed(ctx *gin.Context) {
	idParam := ctx.Param("id")

	var id uint
	if idParam != "" {
		parsed, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, errors.New("некорректный идентификатор услуги"))
			return
		}
		id = uint(parsed)
	}

	next := ctx.Query("next") == "true"

	utility, err := h.Repository.GetUtilityFeedItem(id, next)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, errors.New("услуга не найдена"))
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

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": ds.UtilityFeedItem{
			Utility:    utility,
			Liked:      liked,
			LikesCount: likes,
		},
	})
}

// APIGetDraft — GET /api/utilities/draft: черновик текущего пользователя
// (не более одного), ID не указывается. Отсутствие черновика — не ошибка,
// data будет null.
func (h *Handler) APIGetDraft(ctx *gin.Context) {
	draft, err := h.Repository.GetUtilityDraft(CurrentUserID())
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": draft})
}

// APIGetUtility — GET /api/utilities/:id: подробная информация об
// опубликованной услуге с вложенным создателем (Preload).
func (h *Handler) APIGetUtility(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("некорректный идентификатор услуги"))
		return
	}

	utility, err := h.Repository.GetUtility(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, errors.New("услуга не найдена"))
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	full := ds.FullUtilitySerializer{Utility: utility, Creator: utility.Creator}
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": full})
}

// APICreateUtility — POST /api/utilities: создаёт черновик услуги по
// названию и, если переданы, загружает изображение и видео как файлы
// (multipart/form-data) в MinIO — в отличие от SSR-формы лабы 2, файлы
// здесь реально сохраняются (требование методички лабораторной 3).
func (h *Handler) APICreateUtility(ctx *gin.Context) {
	if err := ctx.Request.ParseMultipartForm(maxUploadSize); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	name := ctx.Request.FormValue("name")
	if name == "" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("поле name обязательно"))
		return
	}

	draft, err := h.Repository.CreateUtilityDraft(CurrentUserID(), name)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if imageHeader, err := ctx.FormFile("image"); err == nil {
		url, err := h.Repository.UploadUtilityImage(draft.ID, imageHeader)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
		draft.ImageURL = url
	} else if !errors.Is(err, http.ErrMissingFile) {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if videoHeader, err := ctx.FormFile("video"); err == nil {
		url, err := h.Repository.UploadUtilityVideo(draft.ID, videoHeader)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
		draft.VideoURL = url
	} else if !errors.Is(err, http.ErrMissingFile) {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"data":    draft,
		"message": "услуга успешно добавлена",
	})
}

// publishRequest — тело PUT-запроса публикации услуги.
type publishRequest struct {
	Description     string  `json:"description"`
	TariffRate      float64 `json:"tariffRate"`
	ConsumptionNorm float64 `json:"consumptionNorm"`
	MeasurementUnit string  `json:"measurementUnit"`
}

// APIPublishUtility — PUT /api/utilities/:id: заполняет оставшиеся поля
// черновика и меняет статус на «опубликован». Один из двух методов смены
// статуса у создателя — вернуть в черновик нельзя.
func (h *Handler) APIPublishUtility(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("некорректный идентификатор услуги"))
		return
	}

	var req publishRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.PublishUtility(uint(id), req.Description, req.TariffRate, req.ConsumptionNorm, req.MeasurementUnit)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, errors.New("черновик не найден"))
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "услуга опубликована"})
}

// APIDeleteUtility — DELETE /api/utilities/:id: логическое удаление,
// только если услуга принадлежит текущему пользователю.
func (h *Handler) APIDeleteUtility(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("некорректный идентификатор услуги"))
		return
	}

	if err := h.Repository.DeleteUtilityOwned(uint(id), CurrentUserID()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, errors.New("услуга не найдена или принадлежит другому пользователю"))
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "услуга удалена"})
}

// likeRequest — тело запроса лайка.
type likeRequest struct {
	Value int `json:"value"` // 0 — отменить лайк, 1 — поставить
}

// APISetLike — POST /api/utilities/:id/like: ставит или снимает лайк
// текущего пользователя.
func (h *Handler) APISetLike(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("некорректный идентификатор услуги"))
		return
	}

	var req likeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.SetLike(CurrentUserID(), uint(id), req.Value); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	likes, err := h.Repository.GetLikesCount(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": gin.H{"likesCount": likes}})
}

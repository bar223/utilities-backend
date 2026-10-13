package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"utility-bills-backend/internal/app/ds"
)

// Изображение и видео по умолчанию хранятся прямо на SSR-сервере (не в
// MinIO) и подставляются, если у услуги нет своего файла.
const (
	defaultImageURL = "/static/img/default-image.svg"
	defaultVideoURL = "/static/video/default-video.mp4"
)

// sliderUnlimited — крайнее положение слайдера фильтра на плитке,
// означающее "без ограничения по тарифу".
const sliderUnlimited = "101"

// utilityView — данные услуги для шаблонов: сама запись плюс количество
// лайков (вычисляется по таблице likes) и изображение/видео с фолбэком на
// значения по умолчанию.
type utilityView struct {
	ds.Utility
	LikesCount      int64
	DisplayImageURL string
	DisplayVideoURL string
}

func (h *Handler) toView(u ds.Utility) utilityView {
	likes, err := h.Repository.GetLikesCount(u.ID)
	if err != nil {
		logrus.Error(err)
	}

	imageURL := u.ImageURL
	if imageURL == "" {
		imageURL = defaultImageURL
	}
	videoURL := u.VideoURL
	if videoURL == "" {
		videoURL = defaultVideoURL
	}

	return utilityView{
		Utility:         u,
		LikesCount:      likes,
		DisplayImageURL: imageURL,
		DisplayVideoURL: videoURL,
	}
}

// Tile — GET /: страница «Плитка», список опубликованных Utility в две
// колонки с фильтром по тарифу.
func (h *Handler) Tile(ctx *gin.Context) {
	tariffQuery := ctx.Query("tariff")

	var maxTariffRate *float64
	// Крайнее положение слайдера (101) — отдельное значение "без
	// ограничения", а не реальный тариф: диапазон 0-100 с шагом 1
	// даёт точную настройку под основную массу тарифов, а это
	// положение отдельно открывает и редкие дорогие услуги.
	if tariffQuery != "" && tariffQuery != sliderUnlimited {
		if parsed, err := strconv.ParseFloat(tariffQuery, 64); err == nil {
			maxTariffRate = &parsed
		} else {
			logrus.Warnf("некорректное значение фильтра tariff=%q: %v", tariffQuery, err)
		}
	}

	utilities, err := h.Repository.GetUtilities(maxTariffRate)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	views := make([]utilityView, 0, len(utilities))
	for _, u := range utilities {
		views = append(views, h.toView(u))
	}

	ctx.HTML(http.StatusOK, "tile.html", gin.H{
		"Utilities": views,
		"Tariff":    tariffQuery,
	})
}

// Feed — GET /feed и GET /feed/:id: страница «Лента», одна услуга на
// экран, ?next=true — переход к следующей.
func (h *Handler) Feed(ctx *gin.Context) {
	idParam := ctx.Param("id")

	var id uint
	if idParam != "" {
		parsed, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			ctx.String(http.StatusBadRequest, "некорректный идентификатор услуги")
			return
		}
		id = uint(parsed)
	}

	next := ctx.Query("next") == "true"

	utility, err := h.Repository.GetUtilityFeedItem(id, next)
	if err != nil {
		ctx.String(http.StatusNotFound, "услуга не найдена")
		return
	}

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"Utility": h.toView(utility),
	})
}

// Add — GET /add: черновик текущего создателя. Если черновика ещё нет,
// показывается форма создания (кнопка «Далее»), если уже есть — форма
// публикации с заполненными полями (кнопка «Опубликовать»).
func (h *Handler) Add(ctx *gin.Context) {
	draft, err := h.Repository.GetUtilityDraft(currentCreatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "add.html", gin.H{
		"Draft": draft,
	})
}

// CreateDraft — POST /add, кнопка «Далее»: создаёт черновик по названию.
// Фото и видео в лабораторной 2 не сохраняются (требование методички) —
// поля выбора файла показаны для полноты интерфейса, но на сервер не
// передаются.
func (h *Handler) CreateDraft(ctx *gin.Context) {
	name := ctx.PostForm("name")

	_, err := h.Repository.CreateUtilityDraft(currentCreatorID, name)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/add")
}

// Publish — POST /publish, кнопка «Опубликовать»: заполняет оставшиеся
// поля черновика и меняет статус на «опубликован».
func (h *Handler) Publish(ctx *gin.Context) {
	idStr := ctx.PostForm("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	description := ctx.PostForm("description")
	measurementUnit := ctx.PostForm("measurementUnit")

	tariffRate, err := strconv.ParseFloat(ctx.PostForm("tariffRate"), 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	consumptionNorm, err := strconv.ParseFloat(ctx.PostForm("consumptionNorm"), 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.PublishUtility(uint(id), description, tariffRate, consumptionNorm, measurementUnit); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}

// Delete — POST /delete-resource: логическое удаление услуги с плитки
// прямым SQL UPDATE.
func (h *Handler) Delete(ctx *gin.Context) {
	idStr := ctx.PostForm("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if err := h.Repository.DeleteUtility(uint(id)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}

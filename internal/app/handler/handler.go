// Package handler содержит обработчики HTTP-запросов домена Utility:
// получают параметры запроса, вызывают репозиторий и передают данные в
// html/template.
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"utilities-backend/internal/app/repository"
)

// currentCreatorID оставлен для совместимости со старыми вызовами —
// новый код использует функцию-синглтон CurrentUserID ниже.
const currentCreatorID uint = 1

// CurrentUserID — функция-синглтон: до появления авторизации (лаба 4)
// пользователь-создатель зафиксирован одним и тем же ID во всех методах.
// Требование методички лабораторной 3 — именно функция, а не голая
// константа, чтобы в будущем заменить реализацию на чтение из JWT/сессии
// в одном месте, не трогая вызывающий код.
func CurrentUserID() uint {
	return currentCreatorID
}

// Handler объединяет обработчики страниц вокруг общего репозитория.
type Handler struct {
	Repository *repository.Repository
}

// NewHandler создаёт обработчик с внедрённым репозиторием.
func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// RegisterHandler регистрирует маршруты домена Utility. Пути вложены под
// /utilities — сущность явно видна в адресе, без общих названий действий
// в одиночку (/add, /publish, /delete).
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/", h.Tile)
	router.GET("/utilities/feed", h.Feed)
	router.GET("/utilities/feed/:id", h.Feed)
	router.GET("/utilities/draft", h.Add)
	router.POST("/utilities/draft", h.CreateDraft)
	router.POST("/utilities/:id/publish", h.Publish)
	router.POST("/utilities/:id/delete", h.Delete)
	router.POST("/utilities/:id/like", h.Like)

	router.GET("/users/register", h.RegisterPage)
	router.POST("/users/register", h.RegisterSubmit)
}

// RegisterAPIHandler регистрирует REST JSON API лабораторной 3 — тестируется
// через Postman/Insomnia, отдельно от HTML-страниц лаб 1-2. По требованию
// методички каждый адрес начинается с /api.
func (h *Handler) RegisterAPIHandler(router *gin.Engine) {
	api := router.Group("/api")

	api.GET("/utilities", h.APIListUtilities)
	api.GET("/utilities/feed", h.APIFeed)
	api.GET("/utilities/feed/:id", h.APIFeed)
	api.GET("/utilities/draft", h.APIGetDraft)
	api.GET("/utilities/:id", h.APIGetUtility)
	api.POST("/utilities", h.APICreateUtility)
	api.PUT("/utilities/:id", h.APIPublishUtility)
	api.DELETE("/utilities/:id", h.APIDeleteUtility)
	api.POST("/utilities/:id/like", h.APISetLike)

	api.POST("/users/register", h.APIRegister)
	api.POST("/users/login", h.APILogin)
	api.POST("/users/logout", h.APILogout)
}

// RegisterStatic регистрирует шаблоны и статику.
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, statusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(statusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

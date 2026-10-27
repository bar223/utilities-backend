// Package handler содержит обработчики REST API веб-сервиса: получают
// параметры запроса, вызывают репозиторий и отдают JSON. HTML-страниц в
// веб-сервисе нет — интерфейс появится в SPA (лабораторные 5-7), а пока
// методы тестируются через Postman.
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"utilities-backend/internal/app/repository"
)

// currentUserID — пользователь-создатель, зафиксированный константой до
// появления авторизации (лабораторная 4).
const currentUserID uint = 1

// CurrentUserID — функция-синглтон: во всех методах текущий пользователь
// берётся только отсюда. В лабораторной 4 реализация заменится на чтение
// из JWT, вызывающий код не изменится.
func CurrentUserID() uint {
	return currentUserID
}

// Handler объединяет обработчики вокруг общего репозитория.
type Handler struct {
	Repository *repository.Repository
}

// NewHandler создаёт обработчик с внедрённым репозиторием.
func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// RegisterHandler регистрирует 10 методов REST API: 7 в домене услуг и 3
// в домене пользователей. Каждый адрес начинается с /api.
func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api")

	utilities := api.Group("/utilities")
	utilities.GET("", h.GetUtilities)
	utilities.GET("/feed", h.GetUtilityFeed)
	utilities.GET("/draft", h.GetUtilityDraft)
	utilities.POST("", h.CreateUtility)
	utilities.PUT("/:id", h.PublishUtility)
	utilities.DELETE("/:id", h.DeleteUtility)
	utilities.POST("/:id/like", h.SetUtilityLike)

	users := api.Group("/users")
	users.POST("/register", h.RegisterUser)
	users.POST("/login", h.LoginUser)
	users.POST("/logout", h.LogoutUser)
}

// errorHandler отвечает ошибкой. Признак успеха/ошибки в тело не пишется —
// его передаёт только код состояния HTTP (200, 201, 400, 404, 409, 500).
func (h *Handler) errorHandler(ctx *gin.Context, statusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(statusCode, gin.H{"error": err.Error()})
}

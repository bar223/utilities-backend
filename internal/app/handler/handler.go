// Package handler содержит обработчики HTTP-запросов домена Utility:
// получают параметры запроса, вызывают репозиторий и передают данные в
// html/template.
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"utility-bills-backend/internal/app/repository"
)

// currentCreatorID — создатель зафиксирован константой до появления
// авторизации (добавится в лабораторной 4).
const currentCreatorID uint = 1

// Handler объединяет обработчики страниц вокруг общего репозитория.
type Handler struct {
	Repository *repository.Repository
}

// NewHandler создаёт обработчик с внедрённым репозиторием.
func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// RegisterHandler регистрирует маршруты домена Utility.
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/", h.Tile)
	router.GET("/feed", h.Feed)
	router.GET("/feed/:id", h.Feed)
	router.GET("/add", h.Add)
	router.POST("/add", h.CreateDraft)
	router.POST("/publish", h.Publish)
	router.POST("/delete-utility", h.Delete)
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

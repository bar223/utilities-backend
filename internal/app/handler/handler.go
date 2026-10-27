// Package handler содержит обработчики REST API веб-сервиса: получают
// параметры запроса, вызывают репозиторий и отдают JSON.
package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"utilities-backend/internal/app/config"
	"utilities-backend/internal/app/redis"
	"utilities-backend/internal/app/repository"
)

// Handler объединяет обработчики вокруг репозитория, конфигурации JWT и
// клиента Redis (чёрный список токенов).
type Handler struct {
	Repository *repository.Repository
	Config     *config.Config
	Redis      *redis.Client
}

// NewHandler создаёт обработчик с внедрёнными зависимостями.
func NewHandler(r *repository.Repository, c *config.Config, rc *redis.Client) *Handler {
	return &Handler{Repository: r, Config: c, Redis: rc}
}

// RegisterHandler регистрирует 10 методов REST API под /api и Swagger UI.
// Метод ленты доступен по двум адресам: без id (панель вкладок) и с id в
// пути. Без авторизации доступно только чтение (список и лента,
// регистрация и вход), методы создателя требуют JWT (WithAuthCheck).
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")

	utilities := api.Group("/utilities")
	utilities.GET("", h.WithOptionalAuth, h.GetUtilities)
	utilities.GET("/feed", h.WithOptionalAuth, h.GetUtilityFeed)
	utilities.GET("/feed/:id", h.WithOptionalAuth, h.GetUtilityFeedByID)
	utilities.GET("/draft", h.WithAuthCheck, h.GetUtilityDraft)
	utilities.POST("", h.WithAuthCheck, h.CreateUtility)
	utilities.PUT("/:id", h.WithAuthCheck, h.PublishUtility)
	utilities.DELETE("/:id", h.WithAuthCheck, h.DeleteUtility)
	utilities.POST("/:id/like", h.WithAuthCheck, h.SetUtilityLike)

	users := api.Group("/users")
	users.POST("/register", h.RegisterUser)
	users.POST("/login", h.LoginUser)
	users.POST("/logout", h.WithAuthCheck, h.LogoutUser)
}

// errorResponse — тело ответа с ошибкой. Признака успеха в теле нет —
// его передаёт только код состояния HTTP.
type errorResponse struct {
	Error string `json:"error" example:"услуга не найдена"`
}

func (h *Handler) errorHandler(ctx *gin.Context, statusCode int, err error) {
	logrus.Error(err.Error())
	ctx.AbortWithStatusJSON(statusCode, errorResponse{Error: err.Error()})
}

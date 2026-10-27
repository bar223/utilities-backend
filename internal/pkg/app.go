// Package pkg собирает приложение из конфигурации, роутера и обработчиков
// и запускает HTTP-сервер.
package pkg

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"utilities-backend/internal/app/config"
	"utilities-backend/internal/app/handler"
)

// Application — собранное приложение: конфигурация, роутер, обработчики.
type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
}

// NewApp создаёт приложение.
func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler) *Application {
	return &Application{Config: c, Router: r, Handler: h}
}

// RunApp регистрирует маршруты и запускает HTTP-сервер.
func (a *Application) RunApp() {
	logrus.Info("Server start up")

	a.Handler.RegisterHandler(a.Router)
	a.Handler.RegisterStatic(a.Router)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}

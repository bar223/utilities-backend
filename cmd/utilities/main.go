// Точка входа приложения. По методическим указаниям курса файл main.go
// должен содержать минимум логики — сборка и запуск вынесены в
// internal/pkg.
package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	_ "utilities-backend/docs/swagger"
	"utilities-backend/internal/app/config"
	"utilities-backend/internal/app/dsn"
	"utilities-backend/internal/app/handler"
	"utilities-backend/internal/app/redis"
	"utilities-backend/internal/app/repository"
	"utilities-backend/internal/pkg"
)

// @title           Utilities API — квитанция по квартплате
// @version         1.0
// @description     Веб-сервис коммунальных услуг: список и лента услуг, черновик, публикация, лайки, регистрация и JWT-аутентификация.
// @host            localhost:8090
// @BasePath        /api
// @schemes         http
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Введите "Bearer <токен из /users/login>"
func main() {
	logrus.Info("Application start!")

	router := gin.Default()

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	repo, err := repository.New(repository.SettingsFromEnv(dsn.FromEnv()))
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	redisClient, err := redis.New(context.Background(), conf.Redis)
	if err != nil {
		logrus.Fatalf("error connecting to redis: %v", err)
	}

	h := handler.NewHandler(repo, conf, redisClient)

	application := pkg.NewApp(conf, router, h)
	application.RunApp()

	logrus.Info("Application terminated!")
}

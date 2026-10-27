// Точка входа приложения. По методическим указаниям курса файл main.go
// должен содержать минимум логики — сборка и запуск вынесены в
// internal/pkg.
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"utilities-backend/internal/app/config"
	"utilities-backend/internal/app/dsn"
	"utilities-backend/internal/app/handler"
	"utilities-backend/internal/app/repository"
	"utilities-backend/internal/pkg"
)

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

	h := handler.NewHandler(repo)

	application := pkg.NewApp(conf, router, h)
	application.RunApp()

	logrus.Info("Application terminated!")
}

// Утилита миграций: создаёт схему БД по моделям пакета ds.
// Запуск: go run ./cmd/migrate
package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"utility-bills-backend/internal/app/ds"
	"utility-bills-backend/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(
		&ds.User{},
		&ds.Utility{},
		&ds.Like{},
	)
	if err != nil {
		panic("cant migrate db: " + err.Error())
	}
}

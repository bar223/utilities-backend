// Утилита миграций: создаёт схему БД по моделям пакета ds.
// Запуск: go run ./cmd/migrate
package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"utilities-backend/internal/app/ds"
	"utilities-backend/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// image_url, video_url и measurement_unit стали NOT NULL: пустые
	// значения в уже существующих строках заменяются пустой строкой, иначе
	// AutoMigrate не сможет добавить ограничение.
	if db.Migrator().HasTable(&ds.Utility{}) {
		for _, column := range []string{"image_url", "video_url", "measurement_unit"} {
			db.Exec("UPDATE utilities SET " + column + " = '' WHERE " + column + " IS NULL")
		}
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

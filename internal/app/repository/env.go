package repository

import "os"

// SettingsFromEnv собирает Settings из переменных окружения (DSN
// PostgreSQL + параметры подключения к MinIO).
func SettingsFromEnv(postgresDSN string) *Settings {
	return &Settings{
		PostgresDSN: postgresDSN,

		MinioEndpoint:   os.Getenv("MINIO_ENDPOINT") + ":" + os.Getenv("MINIO_PORT"),
		MinioAccessKey:  os.Getenv("MINIO_ROOT_USER"),
		MinioSecretKey:  os.Getenv("MINIO_ROOT_PASSWORD"),
		MinioBucketName: os.Getenv("MINIO_BUCKET"),
		MinioUseSSL:     os.Getenv("MINIO_USE_SSL") == "true",
	}
}

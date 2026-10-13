// Package repository отвечает за доступ к данным — подключение к
// PostgreSQL через GORM и к MinIO для файлов услуг.
package repository

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Settings — параметры подключения репозитория к внешним хранилищам.
type Settings struct {
	PostgresDSN string

	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucketName string
	MinioUseSSL     bool
}

// Repository хранит подключение к БД и к MinIO.
type Repository struct {
	db              *gorm.DB
	minio           *minio.Client
	minioEndpoint   string
	minioBucketName string
}

// New открывает подключение к PostgreSQL и к MinIO по переданным
// настройкам.
func New(s *Settings) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(s.PostgresDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	minioClient, err := minio.New(s.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.MinioAccessKey, s.MinioSecretKey, ""),
		Secure: s.MinioUseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:              db,
		minio:           minioClient,
		minioEndpoint:   s.MinioEndpoint,
		minioBucketName: s.MinioBucketName,
	}, nil
}

package repository

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/minio/minio-go/v7"

	"utility-bills-backend/internal/app/ds"
)

// minioPublicURL собирает публичную ссылку на объект в бакете MinIO.
func minioPublicURL(endpoint, bucket, object string) string {
	return "http://" + endpoint + "/" + bucket + "/" + object
}

// uploadFile загружает файл в MinIO под именем на латинице
// ("utility-<id>-<kind><расширение>") и возвращает публичную ссылку.
func (r *Repository) uploadFile(utilityID uint, kind string, header *multipart.FileHeader) (string, error) {
	file, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	// Тип содержимого определяется по первым байтам файла, а не по
	// заголовку, присланному клиентом — ему доверять нельзя.
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return "", fmt.Errorf("ошибка чтения файла: %w", err)
	}
	contentType := http.DetectContentType(buffer[:n])

	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("ошибка перемещения по файловому потоку: %w", err)
	}

	objectName := "utility-" + strconv.FormatUint(uint64(utilityID), 10) + "-" + kind + filepath.Ext(header.Filename)

	ctx := context.Background()
	_, err = r.minio.PutObject(ctx, r.minioBucketName, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("не удалось добавить объект в хранилище MinIO: %w", err)
	}

	return minioPublicURL(r.minioEndpoint, r.minioBucketName, objectName), nil
}

// UploadUtilityImage загружает изображение услуги в MinIO и сохраняет
// ссылку в БД.
func (r *Repository) UploadUtilityImage(utilityID uint, header *multipart.FileHeader) (string, error) {
	url, err := r.uploadFile(utilityID, "image", header)
	if err != nil {
		return "", err
	}
	if err := r.db.Model(&ds.Utility{}).Where("id = ?", utilityID).UpdateColumn("image_url", url).Error; err != nil {
		return "", fmt.Errorf("ошибка сохранения ссылки на изображение: %w", err)
	}
	return url, nil
}

// UploadUtilityVideo загружает видео услуги в MinIO и сохраняет ссылку
// в БД.
func (r *Repository) UploadUtilityVideo(utilityID uint, header *multipart.FileHeader) (string, error) {
	url, err := r.uploadFile(utilityID, "video", header)
	if err != nil {
		return "", err
	}
	if err := r.db.Model(&ds.Utility{}).Where("id = ?", utilityID).UpdateColumn("video_url", url).Error; err != nil {
		return "", fmt.Errorf("ошибка сохранения ссылки на видео: %w", err)
	}
	return url, nil
}

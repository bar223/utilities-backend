package repository

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/minio/minio-go/v7"

	"utilities-backend/internal/app/ds"
)

// ErrBadFileType — тип загруженного файла не совпадает с ожидаемым
// (в поле image не изображение или в поле video не видео).
var ErrBadFileType = errors.New("неверный тип файла")

// minioPublicURL собирает публичную ссылку на объект в бакете MinIO.
func minioPublicURL(endpoint, bucket, object string) string {
	return "http://" + endpoint + "/" + bucket + "/" + object
}

// DetectFileType определяет тип содержимого файла по первым байтам, а не
// по заголовку, присланному клиентом (ему доверять нельзя), и проверяет,
// что это kind ("image" или "video"). Иначе — ErrBadFileType.
func DetectFileType(kind string, header *multipart.FileHeader) (string, error) {
	file, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return "", fmt.Errorf("ошибка чтения файла: %w", err)
	}
	// Сигнатуры некоторых форматов (например, AVIF) стандартная библиотека
	// не знает — тогда тип берётся по расширению файла.
	contentType := http.DetectContentType(buffer[:n])
	if !strings.HasPrefix(contentType, kind+"/") {
		contentType = mime.TypeByExtension(strings.ToLower(filepath.Ext(header.Filename)))
	}
	if !strings.HasPrefix(contentType, kind+"/") {
		return "", fmt.Errorf("%w: ожидается %s, получен %s", ErrBadFileType, kind, contentType)
	}
	return contentType, nil
}

// uploadFile загружает файл в MinIO под именем на латинице
// ("utility-<id>-<kind><расширение>") и возвращает публичную ссылку.
func (r *Repository) uploadFile(utilityID uint, kind string, header *multipart.FileHeader) (string, error) {
	contentType, err := DetectFileType(kind, header)
	if err != nil {
		return "", err
	}

	file, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

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

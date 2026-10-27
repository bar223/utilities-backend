package repository

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"

	"utilities-backend/internal/app/ds"
)

// GetUtilities — список опубликованных услуг для страницы «Плитка»,
// отсортированный по ID. Если maxTariffRate указан, остаются только услуги
// с тарифом не выше этого значения (фильтр по теме).
func (r *Repository) GetUtilities(maxTariffRate *float64) ([]ds.Utility, error) {
	query := r.db.Where("status = ?", ds.StatusPublished).Order("id")
	if maxTariffRate != nil {
		query = query.Where("tariff_rate <= ?", *maxTariffRate)
	}

	var utilities []ds.Utility
	if err := query.Find(&utilities).Error; err != nil {
		return nil, err
	}
	return utilities, nil
}

// GetUtilityFeedItem — услуга для страницы «Лента». БД возвращает сразу
// одну строку (не более одного дополнительного запроса при переходе через
// границу последней услуги) — без выборки всего списка и фильтрации
// массивом в коде, как того требует методичка.
//
// Если next равен false, возвращается сама услуга по id (id == 0 — первая
// опубликованная, для перехода из панели вкладок). Если next равен true —
// следующая после неё опубликованная услуга, после последней — снова первая.
func (r *Repository) GetUtilityFeedItem(id uint, next bool) (ds.Utility, error) {
	var utility ds.Utility

	switch {
	case id == 0:
		// Переход из панели вкладок — первая опубликованная услуга.
		err := r.db.Where("status = ?", ds.StatusPublished).
			Order("id").First(&utility).Error
		return utility, err

	case next:
		// Следующая после id услуга одним запросом.
		err := r.db.Where("status = ? AND id > ?", ds.StatusPublished, id).
			Order("id").First(&utility).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// После последней услуги — зацикливаемся на первую.
			err = r.db.Where("status = ?", ds.StatusPublished).
				Order("id").First(&utility).Error
		}
		return utility, err

	default:
		err := r.db.Where("status = ? AND id = ?", ds.StatusPublished, id).
			First(&utility).Error
		return utility, err
	}
}

// GetUtilityDraft — черновик текущего создателя (не более одного на
// пользователя). Отсутствие черновика — не ошибка, возвращается nil.
func (r *Repository) GetUtilityDraft(creatorID uint) (*ds.Utility, error) {
	var draft ds.Utility
	err := r.db.Where("status = ? AND creator_id = ?", ds.StatusDraft, creatorID).First(&draft).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &draft, nil
}

// CreateUtilityDraft создаёт черновик услуги по кнопке «Далее». Фото и
// видео в лабораторной 2 не передаются на сервер и не сохраняются
// (требование методички) — вместо них в обязательные поля image_url и
// video_url записываются заведомо неверные ссылки на MinIO. Файлов по ним
// нет, поэтому на страницах показываются фото и видео по умолчанию.
func (r *Repository) CreateUtilityDraft(creatorID uint, name string) (*ds.Utility, error) {
	draft := ds.Utility{
		Name:      name,
		Status:    ds.StatusDraft,
		CreatorID: creatorID,
		CreatedAt: time.Now(),
	}
	if err := r.db.Create(&draft).Error; err != nil {
		return nil, err
	}

	draft.ImageURL = missingMinioURL(draft.ID, "image.jpg")
	draft.VideoURL = missingMinioURL(draft.ID, "video.mp4")
	err := r.db.Model(&draft).Updates(map[string]any{
		"image_url": draft.ImageURL,
		"video_url": draft.VideoURL,
	}).Error
	if err != nil {
		return nil, err
	}
	return &draft, nil
}

// missingMinioURL — ссылка на несуществующий объект в бакете MinIO
// (латинское имя "utility-<id>-<suffix>").
func missingMinioURL(id uint, suffix string) string {
	return fmt.Sprintf("http://%s:%s/%s/utility-%d-%s",
		os.Getenv("MINIO_ENDPOINT"), os.Getenv("MINIO_PORT"), os.Getenv("MINIO_BUCKET"), id, suffix)
}

// PublishUtility заполняет оставшиеся поля черновика и переводит его в
// статус «опубликован» по кнопке «Опубликовать».
func (r *Repository) PublishUtility(id uint, description string, tariffRate, consumptionNorm float64, measurementUnit string) error {
	now := time.Now()
	return r.db.Model(&ds.Utility{}).
		Where("id = ? AND status = ?", id, ds.StatusDraft).
		Updates(map[string]any{
			"description":      description,
			"tariff_rate":      tariffRate,
			"consumption_norm": consumptionNorm,
			"measurement_unit": measurementUnit,
			"status":           ds.StatusPublished,
			"published_at":     now,
		}).Error
}

// DeleteUtility — логическое удаление услуги прямым SQL UPDATE, без ORM
// (требование методички лабораторной 2).
func (r *Repository) DeleteUtility(id uint) error {
	return r.db.Exec("UPDATE utilities SET status = ? WHERE id = ?", ds.StatusDeleted, id).Error
}

// GetLikesCount возвращает количество лайков услуги — вычисляется по
// таблице likes, а не хранится отдельным полем.
func (r *Repository) GetLikesCount(utilityID uint) (int64, error) {
	var count int64
	err := r.db.Model(&ds.Like{}).Where("utility_id = ?", utilityID).Count(&count).Error
	return count, err
}

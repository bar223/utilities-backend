package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"utilities-backend/internal/app/ds"
)

// ErrDraftExists возвращается при попытке создать второй черновик — у
// пользователя не более одной услуги в статусе «черновик».
var ErrDraftExists = errors.New("у пользователя уже есть черновик")

// GetUtilities — список опубликованных услуг, отсортированный по ID. Если
// maxTariffRate указан, остаются только услуги с тарифом не выше этого
// значения (фильтрация на бэкенде).
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

// GetUtilityFeedItem — одна опубликованная услуга для ленты вместе с
// создателем (Preload). БД возвращает сразу одну строку.
//
// id == 0 — первая опубликованная услуга (переход из панели вкладок, id
// не указывается). next == false — сама услуга по id, next == true —
// следующая после неё, после последней — снова первая.
func (r *Repository) GetUtilityFeedItem(id uint, next bool) (ds.Utility, error) {
	var utility ds.Utility
	// каждый вызов — новый запрос, чтобы условия не накапливались
	published := func() *gorm.DB {
		return r.db.Preload("Creator").Where("status = ?", ds.StatusPublished).Order("id")
	}

	switch {
	case id == 0:
		err := published().First(&utility).Error
		return utility, err

	case next:
		err := published().Where("id > ?", id).First(&utility).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = published().First(&utility).Error
		}
		return utility, err

	default:
		err := published().Where("id = ?", id).First(&utility).Error
		return utility, err
	}
}

// GetUtilityDraft — черновик пользователя (не более одного).
func (r *Repository) GetUtilityDraft(creatorID uint) (ds.Utility, error) {
	var draft ds.Utility
	err := r.db.Where("status = ? AND creator_id = ?", ds.StatusDraft, creatorID).First(&draft).Error
	return draft, err
}

// CreateUtilityDraft создаёт черновик услуги по названию. Если черновик у
// пользователя уже есть, возвращает ErrDraftExists.
func (r *Repository) CreateUtilityDraft(creatorID uint, name string) (*ds.Utility, error) {
	_, err := r.GetUtilityDraft(creatorID)
	if err == nil {
		return nil, ErrDraftExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	draft := ds.Utility{
		Name:      name,
		Status:    ds.StatusDraft,
		CreatorID: creatorID,
		CreatedAt: time.Now(),
	}
	if err := r.db.Create(&draft).Error; err != nil {
		return nil, err
	}
	return &draft, nil
}

// PublishUtility заполняет оставшиеся поля черновика создателя и переводит
// его в статус «опубликован». Условие status = draft пропускает только
// переход черновик → опубликован, вернуть в черновик нельзя. Возвращает
// обновлённую запись.
func (r *Repository) PublishUtility(id, creatorID uint, req ds.PublishUtilityRequest) (ds.Utility, error) {
	var utility ds.Utility
	res := r.db.Model(&ds.Utility{}).
		Where("id = ? AND creator_id = ? AND status = ?", id, creatorID, ds.StatusDraft).
		Updates(map[string]any{
			"description":      req.Description,
			"tariff_rate":      req.TariffRate,
			"consumption_norm": req.ConsumptionNorm,
			"measurement_unit": req.MeasurementUnit,
			"status":           ds.StatusPublished,
			"published_at":     time.Now(),
		})
	if res.Error != nil {
		return utility, res.Error
	}
	if res.RowsAffected == 0 {
		return utility, gorm.ErrRecordNotFound
	}

	err := r.db.First(&utility, id).Error
	return utility, err
}

// DeleteUtility — логическое удаление (status = deleted) через ORM, только
// если услуга принадлежит указанному создателю и ещё не удалена.
func (r *Repository) DeleteUtility(id, creatorID uint) error {
	res := r.db.Model(&ds.Utility{}).
		Where("id = ? AND creator_id = ? AND status <> ?", id, creatorID, ds.StatusDeleted).
		Update("status", ds.StatusDeleted)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// IsUtilityPublished проверяет, что услуга существует и опубликована —
// лайкать черновики и удалённые услуги нельзя.
func (r *Repository) IsUtilityPublished(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&ds.Utility{}).Where("id = ? AND status = ?", id, ds.StatusPublished).Count(&count).Error
	return count > 0, err
}

// GetLikesCount возвращает количество лайков услуги — вычисляется по
// таблице likes, а не хранится отдельным полем.
func (r *Repository) GetLikesCount(utilityID uint) (int64, error) {
	var count int64
	err := r.db.Model(&ds.Like{}).Where("utility_id = ?", utilityID).Count(&count).Error
	return count, err
}

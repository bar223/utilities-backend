package repository

import (
	"errors"

	"gorm.io/gorm"

	"utility-bills-backend/internal/app/ds"
)

// SetLike ставит (value=1) или снимает (value=0) лайк пользователя на
// услуге. Идемпотентно: повторная установка уже стоящего лайка не
// создаёт дубликат (есть уникальный индекс на паре user_id/utility_id).
func (r *Repository) SetLike(userID, utilityID uint, value int) error {
	if value == 0 {
		return r.db.Where("user_id = ? AND utility_id = ?", userID, utilityID).Delete(&ds.Like{}).Error
	}

	var existing ds.Like
	err := r.db.Where("user_id = ? AND utility_id = ?", userID, utilityID).First(&existing).Error
	if err == nil {
		return nil // лайк уже стоит
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	like := ds.Like{UserID: userID, UtilityID: utilityID}
	return r.db.Create(&like).Error
}

// IsLikedByUser проверяет, лайкнул ли пользователь услугу.
func (r *Repository) IsLikedByUser(userID, utilityID uint) (bool, error) {
	var count int64
	err := r.db.Model(&ds.Like{}).Where("user_id = ? AND utility_id = ?", userID, utilityID).Count(&count).Error
	return count > 0, err
}

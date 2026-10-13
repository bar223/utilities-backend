package ds

// Like — связь м-м «пользователь — услуга»: отметка «нравится» конкретного
// пользователя на конкретной услуге. Таблица likes.
type Like struct {
	ID uint `gorm:"primaryKey"`

	UserID    uint `gorm:"not null;uniqueIndex:idx_user_utility"`
	UtilityID uint `gorm:"not null;uniqueIndex:idx_user_utility"`

	User    User    `gorm:"foreignKey:UserID"`
	Utility Utility `gorm:"foreignKey:UtilityID"`
}

func (Like) TableName() string {
	return "likes"
}

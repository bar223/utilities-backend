package ds

import "time"

// Статусы Utility — жизненный цикл записи справочника услуг.
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusDeleted   = "deleted"
)

// Utility — коммунальный ресурс (услуга), основная сущность варианта
// «Квитанция по квартплате». Таблица utilities.
type Utility struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Name            string  `gorm:"type:varchar(100);not null" json:"name"`
	Description     string  `gorm:"type:varchar(500)" json:"description"`
	Status          string  `gorm:"type:varchar(15);not null;default:draft" json:"status"`
	ImageURL        string  `gorm:"type:varchar(255)" json:"imageUrl"`
	VideoURL        string  `gorm:"type:varchar(255)" json:"videoUrl"`
	TariffRate      float64 `gorm:"type:numeric(10,2)" json:"tariffRate"`
	ConsumptionNorm float64 `gorm:"type:numeric(10,3)" json:"consumptionNorm"` // поле по теме
	MeasurementUnit string  `gorm:"type:varchar(20)" json:"measurementUnit"`

	CreatedAt   time.Time  `json:"createdAt" gorm:"not null"`
	PublishedAt *time.Time `json:"publishedAt"` // дата формирования (публикации)

	CreatorID uint `gorm:"not null" json:"-"`
	Creator   User `gorm:"foreignKey:CreatorID" json:"-"`
}

func (Utility) TableName() string {
	return "utilities"
}

// FullUtilitySerializer — подробное представление услуги с вложенным
// создателем (вложенная сериализация). Используется для GET одной услуги.
type FullUtilitySerializer struct {
	Utility
	Creator User `json:"creator"`
}

// UtilityListItem — услуга в списке (плитка): добавлен признак "это моя
// услуга" (создатель совпадает с текущим пользователем) и количество
// лайков.
type UtilityListItem struct {
	Utility
	IsMine     bool  `json:"isMine"`
	LikesCount int64 `json:"likesCount"`
}

// UtilityFeedItem — услуга в ленте: вместо признака "моя" добавлен
// признак "лайкнул ли текущий пользователь".
type UtilityFeedItem struct {
	Utility
	Liked      bool  `json:"liked"`
	LikesCount int64 `json:"likesCount"`
}

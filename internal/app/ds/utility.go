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
	ImageURL        string  `gorm:"type:varchar(255);not null" json:"imageUrl"`
	VideoURL        string  `gorm:"type:varchar(255);not null" json:"videoUrl"`
	TariffRate      float64 `gorm:"type:numeric(10,2)" json:"tariffRate"`
	ConsumptionNorm float64 `gorm:"type:numeric(10,3)" json:"consumptionNorm"` // поле по теме
	// Единица измерения обязательна, но формат создания черновика не
	// меняется: при создании записывается пустая строка, реальное значение
	// приходит при публикации.
	MeasurementUnit string `gorm:"type:varchar(20);not null;default:''" json:"measurementUnit"`

	CreatedAt   time.Time  `json:"createdAt" gorm:"not null"`
	PublishedAt *time.Time `json:"publishedAt"` // дата формирования (публикации)

	CreatorID uint `gorm:"not null" json:"-"`
	Creator   User `gorm:"foreignKey:CreatorID" json:"-"`
}

func (Utility) TableName() string {
	return "utilities"
}

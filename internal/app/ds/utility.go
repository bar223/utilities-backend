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
	ID uint `gorm:"primaryKey"`

	Name            string  `gorm:"type:varchar(100);not null"`
	Description     string  `gorm:"type:varchar(500)"`
	Status          string  `gorm:"type:varchar(15);not null;default:draft"`
	ImageURL        string  `gorm:"type:varchar(255);not null"`
	VideoURL        string  `gorm:"type:varchar(255);not null"`
	TariffRate      float64 `gorm:"type:numeric(10,2)"`
	ConsumptionNorm float64 `gorm:"type:numeric(10,3)"` // норматив потребления — поле по теме
	// Единица измерения обязательна, но формат создания черновика не
	// меняется: по «Далее» записывается пустая строка, реальное значение
	// приходит при публикации.
	MeasurementUnit string `gorm:"type:varchar(20);not null;default:''"`

	CreatedAt   time.Time  `gorm:"not null"`
	PublishedAt *time.Time // дата формирования (публикации)

	CreatorID uint `gorm:"not null"`
	Creator   User `gorm:"foreignKey:CreatorID"`
}

func (Utility) TableName() string {
	return "utilities"
}

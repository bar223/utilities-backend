package ds

// Attribute — название дополнительной характеристики услуги.
type Attribute struct {
	ID     uint             `gorm:"primaryKey" json:"id"`
	Name   string           `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Values []AttributeValue `gorm:"foreignKey:AttributeID" json:"-"`
}

func (Attribute) TableName() string { return "attributes" }

// AttributeValue — значение характеристики конкретной услуги.
type AttributeValue struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UtilityID   uint      `gorm:"not null;index:idx_utility_attribute" json:"-"`
	AttributeID uint      `gorm:"not null;index:idx_utility_attribute" json:"-"`
	Value       string    `gorm:"type:varchar(255);not null" json:"value"`
	Attribute   Attribute `gorm:"foreignKey:AttributeID" json:"-"`
}

func (AttributeValue) TableName() string { return "attribute_values" }

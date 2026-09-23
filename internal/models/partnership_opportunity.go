package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PartnershipOpportunity is a specific offer inside a config.
// Example: "₦500k cash sponsorship, 5 brand slots available"
type PartnershipOpportunity struct {
	BaseModel
	EventID uuid.UUID `gorm:"type:uuid;not null;index" json:"event_id"`

	Title       string `gorm:"not null;size:255" json:"title"`
	Description string `gorm:"type:text" json:"description"`

	Type PartnershipType `gorm:"size:40;not null;index" json:"type"`

	// Optional budget range in minor units (currency inherited from event)
	BudgetMinMinor int64 `gorm:"default:0" json:"budget_min_minor"`
	BudgetMaxMinor int64 `gorm:"default:0" json:"budget_max_minor"`

	SlotsTotal     int `gorm:"default:1" json:"slots_total"`
	SlotsRemaining int `gorm:"default:1" json:"slots_remaining"`

	SortOrder int  `gorm:"default:0" json:"sort_order"`
	IsActive  bool `gorm:"default:true" json:"is_active"`

	Event Event `gorm:"foreignKey:EventID" json:"-"`
}

func (o *PartnershipOpportunity) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	if o.SlotsRemaining == 0 {
		o.SlotsRemaining = o.SlotsTotal
	}
	return nil
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketType struct {
	BaseModel
	EventID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"event_id"`
	OccurrenceID *uuid.UUID `gorm:"type:uuid;index" json:"occurrence_id,omitempty"` // optional (per-session tier)

	Name        string `gorm:"not null;size:255" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	ImageURL    string `gorm:"size:500" json:"image_url"`

	PriceMinor int64  `gorm:"not null;default:0" json:"price_minor"`
	Currency   string `gorm:"size:3;not null" json:"currency"`

	QuantityTotal    int `gorm:"not null;default:0" json:"quantity_total"`
	QuantitySold     int `gorm:"not null;default:0" json:"quantity_sold"`
	QuantityReserved int `gorm:"not null;default:0" json:"quantity_reserved"`

	PerUserLimit int `gorm:"default:0" json:"per_user_limit"` // 0 = unlimited

	SalesStartAt *time.Time `json:"sales_start_at,omitempty"`
	SalesEndAt   *time.Time `json:"sales_end_at,omitempty"`

	IsHidden    bool   `gorm:"default:false" json:"is_hidden"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	AccessLevel string `gorm:"size:30;default:'general'" json:"access_level"` // general | vip | backstage

	RequiresApproval bool `gorm:"default:false" json:"requires_approval"`
	SortOrder        int  `gorm:"default:0" json:"sort_order"`

	Event Event `gorm:"foreignKey:EventID" json:"-"`
}

func (t *TicketType) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// Remaining returns the number of tickets still available to sell.
func (t *TicketType) Remaining() int {
	rem := t.QuantityTotal - t.QuantitySold - t.QuantityReserved
	if rem < 0 {
		return 0
	}
	return rem
}

// IsOnSale returns whether the ticket type is currently purchasable.
func (t *TicketType) IsOnSale(now time.Time) bool {
	if !t.IsActive || t.IsHidden {
		return false
	}
	if t.SalesStartAt != nil && now.Before(*t.SalesStartAt) {
		return false
	}
	if t.SalesEndAt != nil && now.After(*t.SalesEndAt) {
		return false
	}
	return t.Remaining() > 0
}

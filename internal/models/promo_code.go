package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PromoType string

const (
	PromoTypePercentage PromoType = "percentage" // value in bps
	PromoTypeFixed      PromoType = "fixed"      // value in minor units
)

type PromoCode struct {
	BaseModel

	EventID *uuid.UUID `gorm:"type:uuid;index" json:"event_id,omitempty"` // nil = platform-wide

	Code string    `gorm:"uniqueIndex;not null;size:50" json:"code"`
	Type PromoType `gorm:"size:20;not null" json:"type"`

	// If percentage, ValueMinor is interpreted as basis points (500 = 5%).
	// If fixed, ValueMinor is the discount in minor units.
	ValueMinor int64 `gorm:"not null" json:"value_minor"`

	UsageLimit   int `gorm:"default:0" json:"usage_limit"` // 0 = unlimited
	UsageCount   int `gorm:"default:0" json:"usage_count"`
	PerUserLimit int `gorm:"default:0" json:"per_user_limit"` // 0 = unlimited

	ValidFrom *time.Time `json:"valid_from,omitempty"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`

	MinOrderMinor int64 `gorm:"default:0" json:"min_order_minor"`

	// Applicable ticket types (nil = all ticket types for the event)
	ApplicableTicketTypeIDs datatypes.JSON `gorm:"type:jsonb" json:"applicable_ticket_type_ids,omitempty"`

	IsActive bool `gorm:"default:true;index" json:"is_active"`

	CreatedByID uuid.UUID `gorm:"type:uuid;not null" json:"created_by_id"`

	Event Event `gorm:"foreignKey:EventID" json:"-"`
}

func (p *PromoCode) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (p *PromoCode) IsValid(now time.Time) bool {
	if !p.IsActive {
		return false
	}
	if p.ValidFrom != nil && now.Before(*p.ValidFrom) {
		return false
	}
	if p.ValidTo != nil && now.After(*p.ValidTo) {
		return false
	}
	if p.UsageLimit > 0 && p.UsageCount >= p.UsageLimit {
		return false
	}
	return true
}

type PromoCodeRedemption struct {
	BaseModel
	PromoCodeID   uuid.UUID `gorm:"type:uuid;not null;index" json:"promo_code_id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	OrderID       uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	DiscountMinor int64     `gorm:"not null" json:"discount_minor"`

	PromoCode PromoCode `gorm:"foreignKey:PromoCodeID" json:"-"`
	User      User      `gorm:"foreignKey:UserID" json:"-"`
	Order     Order     `gorm:"foreignKey:OrderID" json:"-"`
}

func (r *PromoCodeRedemption) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

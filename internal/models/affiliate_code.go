package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AffiliateCode links a PromoCode to a partnership so brands can track
// clicks/purchases attributed to them.
type AffiliateCode struct {
	BaseModel
	PartnershipID uuid.UUID `gorm:"type:uuid;not null;index" json:"partnership_id"`
	PromoCodeID   uuid.UUID `gorm:"type:uuid;not null;index" json:"promo_code_id"`

	// Code displayed to the brand (may equal PromoCode.Code)
	Code string `gorm:"not null;size:50;uniqueIndex" json:"code"`

	// Attribution counters
	Clicks          int64 `gorm:"default:0" json:"clicks"`
	Conversions     int64 `gorm:"default:0" json:"conversions"`
	RevenueMinor    int64 `gorm:"default:0" json:"revenue_minor"`
	CommissionMinor int64 `gorm:"default:0" json:"commission_minor"`

	Partnership PartnershipRequest `gorm:"foreignKey:PartnershipID" json:"-"`
	PromoCode   PromoCode          `gorm:"foreignKey:PromoCodeID" json:"-"`
}

func (a *AffiliateCode) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

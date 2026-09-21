package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommissionTier defines a percentage rate that applies to organizers whose
// lifetime (or windowed) revenue falls within a range.
//
// Rates are stored in basis points (bps): 1 bps = 0.01%. 500 bps = 5%.
// Revenue thresholds are stored in minor units.
type CommissionTier struct {
	BaseModel
	Name            string `gorm:"size:100;not null" json:"name"`
	MinRevenueMinor int64  `gorm:"not null;default:0" json:"min_revenue_minor"`
	MaxRevenueMinor *int64 `json:"max_revenue_minor,omitempty"`     // nil = no upper bound
	RateBps         int    `gorm:"not null" json:"rate_bps"`        // e.g. 500 = 5%
	Priority        int    `gorm:"default:0;index" json:"priority"` // lower = checked first
	IsActive        bool   `gorm:"default:true" json:"is_active"`
}

func (t *CommissionTier) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// OrganizerCommissionOverride allows Super Admins to force a specific rate
// for an organizer regardless of tiers. Set RateBps = -1 to disable (use tier).
type OrganizerCommissionOverride struct {
	BaseModel
	OrganizerID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"organizer_id"`
	RateBps     int        `gorm:"not null;default:-1" json:"rate_bps"`
	Reason      string     `gorm:"size:500" json:"reason"`
	SetBy       *uuid.UUID `gorm:"type:uuid" json:"set_by,omitempty"`

	Organizer User `gorm:"foreignKey:OrganizerID" json:"-"`
}

func (o *OrganizerCommissionOverride) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

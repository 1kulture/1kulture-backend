package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Setting is a globally configurable key/value pair controlled by Super Admins.
type Setting struct {
	BaseModel
	Key         string     `gorm:"uniqueIndex;not null;size:100" json:"key"`
	Value       string     `gorm:"type:text" json:"value"`
	ValueType   string     `gorm:"size:20;not null;default:'string'" json:"value_type"` // string | int | bool | json | duration
	Category    string     `gorm:"size:50;index" json:"category"`
	Description string     `gorm:"size:500" json:"description"`
	IsPublic    bool       `gorm:"default:false" json:"is_public"` // whether clients can read it
	UpdatedBy   *uuid.UUID `gorm:"type:uuid" json:"updated_by,omitempty"`
}

func (s *Setting) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// KnownSettingKeys are canonical keys used throughout the app.
const (
	SettingEscrowReleaseDays       = "escrow.release_days"
	SettingDefaultCommissionRate   = "commission.default_rate_bps"
	SettingDefaultCurrency         = "platform.default_currency"
	SettingPlatformFeeRate         = "platform.processing_fee_bps"
	SettingRefundPolicyDefaultDays = "refund.default_days"
	SettingRefundPolicyDefaultText = "refund.default_text"
	SettingEventMaxGalleryImages   = "event.max_gallery_images"
	SettingTicketTransferEnabled   = "ticket.transfer_enabled"
)

package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShareChannel string

const (
	ShareChannelLink     ShareChannel = "link"
	ShareChannelWhatsApp ShareChannel = "whatsapp"
	ShareChannelX        ShareChannel = "x"
	ShareChannelFacebook ShareChannel = "facebook"
	ShareChannelEmail    ShareChannel = "email"
	ShareChannelOther    ShareChannel = "other"
)

type EventShare struct {
	BaseModel
	EventID uuid.UUID    `gorm:"type:uuid;not null;index" json:"event_id"`
	UserID  *uuid.UUID   `gorm:"type:uuid;index" json:"user_id,omitempty"` // nullable for guests (public pages)
	Channel ShareChannel `gorm:"size:20;not null" json:"channel"`

	ReferralCode string `gorm:"size:20;index" json:"referral_code"`
	IPAddress    string `gorm:"size:50" json:"ip_address"`
	UserAgent    string `gorm:"size:500" json:"user_agent"`

	Event Event `gorm:"foreignKey:EventID" json:"-"`
}

func (s *EventShare) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

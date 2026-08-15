package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmailVerification struct {
	BaseModel
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Email      string     `gorm:"size:255;not null;index" json:"email"`
	Code       string     `gorm:"size:6;not null" json:"code"`
	ExpiresAt  time.Time  `gorm:"not null;index" json:"expires_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	Attempts   int        `gorm:"default:0" json:"attempts"`
	User       User       `gorm:"foreignKey:UserID" json:"-"`
}

func (ev *EmailVerification) BeforeCreate(tx *gorm.DB) error {
	if ev.ID == uuid.Nil {
		ev.ID = uuid.New()
	}
	return nil
}

func (ev *EmailVerification) IsValid() bool {
	return ev.VerifiedAt == nil && time.Now().Before(ev.ExpiresAt) && ev.Attempts < 5
}

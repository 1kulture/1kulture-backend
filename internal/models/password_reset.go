package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasswordReset struct {
	BaseModel
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string     `gorm:"uniqueIndex;not null;size:500" json:"token"`
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	User      User       `gorm:"foreignKey:UserID" json:"-"`
}

func (pr *PasswordReset) BeforeCreate(tx *gorm.DB) error {
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	return nil
}

func (pr *PasswordReset) IsValid() bool {
	return pr.UsedAt == nil && time.Now().Before(pr.ExpiresAt)
}

package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WaitlistEntry struct {
	BaseModel
	Email    string `gorm:"size:255;not null;index" json:"email"`
	Category string `gorm:"size:100;not null" json:"category"`
	Status   string `gorm:"size:20;default:'pending'" json:"status"`
}

func (w *WaitlistEntry) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	if w.Status == "" {
		w.Status = "pending"
	}
	return nil
}

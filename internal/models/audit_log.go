package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditLog struct {
	BaseModel
	UserID     *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action     string     `gorm:"size:100;not null;index" json:"action"`
	Resource   string     `gorm:"size:100;not null;index" json:"resource"`
	ResourceID string     `gorm:"size:100;index" json:"resource_id,omitempty"`
	IPAddress  string     `gorm:"size:50" json:"ip_address"`
	UserAgent  string     `gorm:"size:500" json:"user_agent"`
	Details    string     `gorm:"type:text" json:"details,omitempty"`
	Status     string     `gorm:"size:20;not null;index" json:"status"`
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleName string

const (
	RoleGuest        RoleName = "guest"
	RoleVendor       RoleName = "vendor"
	RoleEventManager RoleName = "event_manager"
	RoleAdmin        RoleName = "admin"
)

type Role struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;not null;size:50" json:"name"`
	Description string `gorm:"size:255" json:"description"`
	Users       []User `gorm:"many2many:user_roles;" json:"-"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

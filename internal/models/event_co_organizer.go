package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// EventCoOrganizerPermissions is a JSONB structure that grants granular permissions.
// Example: {"can_edit": true, "can_publish": false, "can_scan": true, "can_refund": false}
type EventCoOrganizerRole string

const (
	EventCoOrgRoleCoHost  EventCoOrganizerRole = "co_host"
	EventCoOrgRoleManager EventCoOrganizerRole = "manager"
	EventCoOrgRoleStaff   EventCoOrganizerRole = "staff"
)

type EventCoOrganizer struct {
	BaseModel
	EventID     uuid.UUID            `gorm:"type:uuid;not null;index:idx_event_coorg,unique" json:"event_id"`
	UserID      uuid.UUID            `gorm:"type:uuid;not null;index:idx_event_coorg,unique" json:"user_id"`
	Role        EventCoOrganizerRole `gorm:"size:20;not null;default:'co_host'" json:"role"`
	Permissions datatypes.JSON       `gorm:"type:jsonb" json:"permissions"`

	Event Event `gorm:"foreignKey:EventID" json:"-"`
	User  User  `gorm:"foreignKey:UserID" json:"-"`
}

func (c *EventCoOrganizer) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

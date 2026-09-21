package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventStaffRole string

const (
	EventStaffRoleScanner     EventStaffRole = "scanner"
	EventStaffRoleCheckInLead EventStaffRole = "checkin_lead"
	EventStaffRoleCoordinator EventStaffRole = "coordinator"
)

type EventStaff struct {
	BaseModel
	EventID uuid.UUID      `gorm:"type:uuid;not null;index:idx_event_staff,unique" json:"event_id"`
	UserID  uuid.UUID      `gorm:"type:uuid;not null;index:idx_event_staff,unique" json:"user_id"`
	Role    EventStaffRole `gorm:"size:30;not null;default:'scanner'" json:"role"`

	Event Event `gorm:"foreignKey:EventID" json:"-"`
	User  User  `gorm:"foreignKey:UserID" json:"-"`
}

func (s *EventStaff) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

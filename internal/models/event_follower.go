package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventFollower struct {
	BaseModel
	EventID uuid.UUID `gorm:"type:uuid;not null;index:idx_event_follower,unique" json:"event_id"`
	UserID  uuid.UUID `gorm:"type:uuid;not null;index:idx_event_follower,unique" json:"user_id"`

	NotifyNewSessions bool `gorm:"default:true" json:"notify_new_sessions"`
	NotifyUpdates     bool `gorm:"default:true" json:"notify_updates"`

	Event Event `gorm:"foreignKey:EventID" json:"-"`
	User  User  `gorm:"foreignKey:UserID" json:"-"`
}

func (f *EventFollower) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

type EventOrganizerFollower struct {
	BaseModel
	UserID      uuid.UUID `gorm:"type:uuid;not null;index:idx_org_follower,unique" json:"user_id"`      // follower
	OrganizerID uuid.UUID `gorm:"type:uuid;not null;index:idx_org_follower,unique" json:"organizer_id"` // event manager being followed

	NotifyNewEvents bool `gorm:"default:true" json:"notify_new_events"`

	User      User `gorm:"foreignKey:UserID" json:"-"`
	Organizer User `gorm:"foreignKey:OrganizerID" json:"-"`
}

func (f *EventOrganizerFollower) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventOccurrence struct {
	BaseModel
	EventID uuid.UUID `gorm:"type:uuid;not null;index" json:"event_id"`

	SessionName string    `gorm:"size:255;not null" json:"session_name"`
	StartAt     time.Time `gorm:"not null;index" json:"start_at"`
	EndAt       time.Time `gorm:"not null;index" json:"end_at"`

	VenueName    string `gorm:"size:255" json:"venue_name"`
	VenueAddress string `gorm:"size:500" json:"venue_address"`
	VirtualURL   string `gorm:"size:500" json:"virtual_url"`

	Capacity  *int `json:"capacity,omitempty"`
	SortOrder int  `gorm:"default:0" json:"sort_order"`

	Event Event `gorm:"foreignKey:EventID" json:"-"`
}

func (o *EventOccurrence) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

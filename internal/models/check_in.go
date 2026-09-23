package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CheckInMethod string

const (
	CheckInMethodQR     CheckInMethod = "qr"
	CheckInMethodManual CheckInMethod = "manual"
)

type CheckIn struct {
	BaseModel
	TicketID  uuid.UUID `gorm:"type:uuid;not null;index" json:"ticket_id"`
	EventID   uuid.UUID `gorm:"type:uuid;not null;index" json:"event_id"`
	ScannedBy uuid.UUID `gorm:"type:uuid;not null;index" json:"scanned_by"`

	Method     CheckInMethod `gorm:"size:20;not null;default:'qr'" json:"method"`
	ScannedAt  time.Time     `gorm:"not null;index" json:"scanned_at"`
	DeviceInfo string        `gorm:"size:255" json:"device_info,omitempty"`
	Location   string        `gorm:"size:255" json:"location,omitempty"`
	IPAddress  string        `gorm:"size:50" json:"ip_address,omitempty"`

	Ticket Ticket `gorm:"foreignKey:TicketID" json:"-"`
	Event  Event  `gorm:"foreignKey:EventID" json:"-"`
}

func (c *CheckIn) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.ScannedAt.IsZero() {
		c.ScannedAt = time.Now().UTC()
	}
	return nil
}

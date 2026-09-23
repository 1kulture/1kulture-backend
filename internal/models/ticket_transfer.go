package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketTransferStatus string

const (
	TicketTransferPending   TicketTransferStatus = "pending"
	TicketTransferAccepted  TicketTransferStatus = "accepted"
	TicketTransferDeclined  TicketTransferStatus = "declined"
	TicketTransferCancelled TicketTransferStatus = "cancelled"
	TicketTransferExpired   TicketTransferStatus = "expired"
)

type TicketTransfer struct {
	BaseModel

	TicketID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"ticket_id"`
	FromUserID uuid.UUID  `gorm:"type:uuid;not null;index" json:"from_user_id"`
	ToEmail    string     `gorm:"not null;size:255;index" json:"to_email"`
	ToUserID   *uuid.UUID `gorm:"type:uuid;index" json:"to_user_id,omitempty"` // filled when accepted

	Message string `gorm:"type:text" json:"message,omitempty"`

	Token     string               `gorm:"uniqueIndex;not null;size:100" json:"token"`
	Status    TicketTransferStatus `gorm:"size:20;not null;default:'pending';index" json:"status"`
	ExpiresAt time.Time            `gorm:"not null;index" json:"expires_at"`

	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
	DeclinedAt  *time.Time `json:"declined_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	Ticket   Ticket `gorm:"foreignKey:TicketID" json:"-"`
	FromUser User   `gorm:"foreignKey:FromUserID" json:"-"`
}

func (t *TicketTransfer) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Status == "" {
		t.Status = TicketTransferPending
	}
	return nil
}

func (t *TicketTransfer) IsPending() bool {
	return t.Status == TicketTransferPending && time.Now().Before(t.ExpiresAt)
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketStatus string

const (
	TicketStatusIssued      TicketStatus = "issued"
	TicketStatusValid       TicketStatus = "valid"
	TicketStatusUsed        TicketStatus = "used"
	TicketStatusCancelled   TicketStatus = "cancelled"
	TicketStatusRefunded    TicketStatus = "refunded"
	TicketStatusTransferred TicketStatus = "transferred"
	TicketStatusExpired     TicketStatus = "expired"
)

type Ticket struct {
	BaseModel

	OrderID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"order_id"`
	OrderItemID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"order_item_id"`
	TicketTypeID uuid.UUID  `gorm:"type:uuid;not null;index" json:"ticket_type_id"`
	EventID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"event_id"`
	OccurrenceID *uuid.UUID `gorm:"type:uuid;index" json:"occurrence_id,omitempty"`

	// Owner (may change via transfer)
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`

	Code      string `gorm:"uniqueIndex;not null;size:32" json:"code"`
	QRPayload string `gorm:"type:text" json:"qr_payload"`

	// Holder details (may differ from buyer)
	HolderName  string `gorm:"size:255" json:"holder_name"`
	HolderEmail string `gorm:"size:255" json:"holder_email"`
	HolderPhone string `gorm:"size:30" json:"holder_phone"`

	Status TicketStatus `gorm:"size:30;not null;default:'issued';index" json:"status"`

	IssuedAt     time.Time  `gorm:"not null" json:"issued_at"`
	UsedAt       *time.Time `json:"used_at,omitempty"`
	ScannedBy    *uuid.UUID `gorm:"type:uuid" json:"scanned_by,omitempty"`
	ScanLocation string     `gorm:"size:255" json:"scan_location,omitempty"`

	// Transfer (denormalized for quick lookup)
	TransferCount int `gorm:"default:0" json:"transfer_count"`

	Order      Order      `gorm:"foreignKey:OrderID" json:"-"`
	OrderItem  OrderItem  `gorm:"foreignKey:OrderItemID" json:"-"`
	TicketType TicketType `gorm:"foreignKey:TicketTypeID" json:"-"`
	Event      Event      `gorm:"foreignKey:EventID" json:"-"`
	User       User       `gorm:"foreignKey:UserID" json:"-"`
}

func (t *Ticket) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Status == "" {
		t.Status = TicketStatusIssued
	}
	if t.IssuedAt.IsZero() {
		t.IssuedAt = time.Now().UTC()
	}
	return nil
}

func (t *Ticket) IsUsable() bool {
	switch t.Status {
	case TicketStatusIssued, TicketStatusValid, TicketStatusTransferred:
		return true
	}
	return false
}

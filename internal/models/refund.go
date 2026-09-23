package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type RefundStatus string

const (
	RefundStatusRequested RefundStatus = "requested"
	RefundStatusApproved  RefundStatus = "approved"
	RefundStatusRejected  RefundStatus = "rejected"
	RefundStatusProcessed RefundStatus = "processed"
	RefundStatusFailed    RefundStatus = "failed"
)

type Refund struct {
	BaseModel

	OrderID uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	UserID  uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`

	// List of ticket IDs being refunded (JSON array of uuid strings)
	TicketIDs datatypes.JSON `gorm:"type:jsonb;not null" json:"ticket_ids"`

	AmountMinor int64  `gorm:"not null" json:"amount_minor"`
	Currency    string `gorm:"size:3;not null" json:"currency"`

	Reason string       `gorm:"type:text" json:"reason"`
	Status RefundStatus `gorm:"size:30;not null;default:'requested';index" json:"status"`

	RequestedByID uuid.UUID  `gorm:"type:uuid;not null" json:"requested_by_id"`
	ReviewedByID  *uuid.UUID `gorm:"type:uuid" json:"reviewed_by_id,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	ReviewNote    string     `gorm:"size:500" json:"review_note,omitempty"`

	ProviderRefundReference string     `gorm:"size:100" json:"provider_refund_reference,omitempty"`
	ProcessedAt             *time.Time `json:"processed_at,omitempty"`

	// If escrow was already released to organizer, the refund is debited from their balance.
	DebitedFromOrganizerBalance bool `gorm:"default:false" json:"debited_from_organizer_balance"`

	Order Order `gorm:"foreignKey:OrderID" json:"-"`
}

func (r *Refund) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.Status == "" {
		r.Status = RefundStatusRequested
	}
	return nil
}

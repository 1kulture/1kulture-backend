package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PaymentTransactionStatus string

const (
	PaymentTxnInitiated PaymentTransactionStatus = "initiated"
	PaymentTxnSuccess   PaymentTransactionStatus = "success"
	PaymentTxnFailed    PaymentTransactionStatus = "failed"
	PaymentTxnReversed  PaymentTransactionStatus = "reversed"
	PaymentTxnPending   PaymentTransactionStatus = "pending"
)

// PaymentTransaction records every raw interaction with the payment provider.
// It's an append-only log for audit and reconciliation.
type PaymentTransaction struct {
	BaseModel

	OrderID uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	UserID  uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`

	Provider    string `gorm:"size:30;not null;index" json:"provider"` // paystack | flutterwave | stripe
	Reference   string `gorm:"size:100;not null;uniqueIndex" json:"reference"`
	ProviderRef string `gorm:"size:150;index" json:"provider_ref,omitempty"`

	AmountMinor int64  `gorm:"not null" json:"amount_minor"`
	Currency    string `gorm:"size:3;not null" json:"currency"`

	Status        PaymentTransactionStatus `gorm:"size:20;not null;default:'initiated';index" json:"status"`
	StatusMessage string                   `gorm:"size:500" json:"status_message,omitempty"`

	// Raw provider response for debugging/auditing
	RawRequest  datatypes.JSON `gorm:"type:jsonb" json:"-"`
	RawResponse datatypes.JSON `gorm:"type:jsonb" json:"-"`

	AuthorizedAt *time.Time `json:"authorized_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`

	Order Order `gorm:"foreignKey:OrderID" json:"-"`
	User  User  `gorm:"foreignKey:UserID" json:"-"`
}

func (p *PaymentTransaction) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

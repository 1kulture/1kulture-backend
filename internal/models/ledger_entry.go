package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type LedgerAccount string

const (
	LedgerAccountPlatform  LedgerAccount = "platform"
	LedgerAccountOrganizer LedgerAccount = "organizer"
	LedgerAccountVendor    LedgerAccount = "vendor"
	LedgerAccountEscrow    LedgerAccount = "escrow"
	LedgerAccountFee       LedgerAccount = "fee"
	LedgerAccountRefund    LedgerAccount = "refund"
	LedgerAccountPayout    LedgerAccount = "payout"
)

type LedgerEntryType string

const (
	LedgerEntryCredit LedgerEntryType = "credit"
	LedgerEntryDebit  LedgerEntryType = "debit"
)

// LedgerEntry is an immutable, append-only record of every money movement.
// Corrections are made by adding compensating entries, never by editing.
type LedgerEntry struct {
	BaseModel

	Reference string          `gorm:"not null;size:100;index" json:"reference"`
	Account   LedgerAccount   `gorm:"size:30;not null;index" json:"account"`
	EntryType LedgerEntryType `gorm:"size:10;not null" json:"entry_type"`

	// Owner of the account (organizer_id, vendor_id). Nil for platform accounts.
	OwnerID *uuid.UUID `gorm:"type:uuid;index" json:"owner_id,omitempty"`

	AmountMinor int64  `gorm:"not null" json:"amount_minor"`
	Currency    string `gorm:"size:3;not null" json:"currency"`

	// Polymorphic link to what caused this entry
	RelatedEntityType string    `gorm:"size:50;index" json:"related_entity_type"` // order | refund | payout | ...
	RelatedEntityID   uuid.UUID `gorm:"type:uuid;index" json:"related_entity_id"`

	Description string         `gorm:"size:500" json:"description,omitempty"`
	Meta        datatypes.JSON `gorm:"type:jsonb" json:"meta,omitempty"`
}

func (l *LedgerEntry) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

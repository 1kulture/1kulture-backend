package responses

import (
	"time"

	"github.com/google/uuid"
)

type LedgerEntryResponse struct {
	ID        uuid.UUID  `json:"id"`
	Reference string     `json:"reference"`
	Account   string     `json:"account"`
	EntryType string     `json:"entry_type"`
	OwnerID   *uuid.UUID `json:"owner_id,omitempty"`

	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`

	RelatedEntityType string    `json:"related_entity_type"`
	RelatedEntityID   uuid.UUID `json:"related_entity_id"`

	Description string `json:"description,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

type BalanceResponse struct {
	OwnerID        uuid.UUID `json:"owner_id"`
	Currency       string    `json:"currency"`
	HeldMinor      int64     `json:"held_minor"`      // escrow
	AvailableMinor int64     `json:"available_minor"` // released, ready for payout
	PaidOutMinor   int64     `json:"paid_out_minor"`
}

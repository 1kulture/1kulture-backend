package responses

import (
	"time"

	"github.com/google/uuid"
)

type RefundResponse struct {
	ID      uuid.UUID `json:"id"`
	OrderID uuid.UUID `json:"order_id"`
	UserID  uuid.UUID `json:"user_id"`

	TicketIDs []uuid.UUID `json:"ticket_ids"`

	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`

	Reason string `json:"reason"`
	Status string `json:"status"`

	RequestedByID uuid.UUID  `json:"requested_by_id"`
	ReviewedByID  *uuid.UUID `json:"reviewed_by_id,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	ReviewNote    string     `json:"review_note,omitempty"`

	ProviderRefundReference string     `json:"provider_refund_reference,omitempty"`
	ProcessedAt             *time.Time `json:"processed_at,omitempty"`

	DebitedFromOrganizerBalance bool `json:"debited_from_organizer_balance"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

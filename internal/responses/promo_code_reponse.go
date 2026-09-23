package responses

import (
	"time"

	"github.com/google/uuid"
)

type PromoCodeResponse struct {
	ID         uuid.UUID  `json:"id"`
	EventID    *uuid.UUID `json:"event_id,omitempty"`
	Code       string     `json:"code"`
	Type       string     `json:"type"`
	ValueMinor int64      `json:"value_minor"`

	UsageLimit   int `json:"usage_limit"`
	UsageCount   int `json:"usage_count"`
	PerUserLimit int `json:"per_user_limit"`

	ValidFrom *time.Time `json:"valid_from,omitempty"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`

	MinOrderMinor int64 `json:"min_order_minor"`

	ApplicableTicketTypeIDs []string `json:"applicable_ticket_type_ids,omitempty"`

	IsActive bool `json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PromoValidationResponse struct {
	Code             string `json:"code"`
	DiscountMinor    int64  `json:"discount_minor"`
	NewSubtotalMinor int64  `json:"new_subtotal_minor"`
	Message          string `json:"message"`
}

package responses

import (
	"time"

	"github.com/google/uuid"
)

type CurrencyResponse struct {
	ID             uuid.UUID `json:"id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Symbol         string    `json:"symbol"`
	IsEnabled      bool      `json:"is_enabled"`
	IsDefault      bool      `json:"is_default"`
	MinAmountMinor int64     `json:"min_amount_minor"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CurrencyRequestResponse struct {
	ID             uuid.UUID  `json:"id"`
	RequestedBy    uuid.UUID  `json:"requested_by"`
	RequesterName  string     `json:"requester_name,omitempty"`
	RequesterEmail string     `json:"requester_email,omitempty"`
	CurrencyCode   string     `json:"currency_code"`
	Reason         string     `json:"reason"`
	Status         string     `json:"status"`
	ReviewedBy     *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewNote     string     `json:"review_note,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

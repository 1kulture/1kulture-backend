package responses

import (
	"time"

	"github.com/google/uuid"
)

type TicketTypeResponse struct {
	ID           uuid.UUID  `json:"id"`
	EventID      uuid.UUID  `json:"event_id"`
	OccurrenceID *uuid.UUID `json:"occurrence_id,omitempty"`

	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`

	PriceMinor int64  `json:"price_minor"`
	Currency   string `json:"currency"`

	QuantityTotal     int `json:"quantity_total"`
	QuantitySold      int `json:"quantity_sold"`
	QuantityReserved  int `json:"quantity_reserved"`
	QuantityRemaining int `json:"quantity_remaining"`

	PerUserLimit int `json:"per_user_limit"`

	SalesStartAt *time.Time `json:"sales_start_at,omitempty"`
	SalesEndAt   *time.Time `json:"sales_end_at,omitempty"`

	IsHidden    bool   `json:"is_hidden"`
	IsActive    bool   `json:"is_active"`
	AccessLevel string `json:"access_level"`

	RequiresApproval bool `json:"requires_approval"`
	SortOrder        int  `json:"sort_order"`

	OnSale bool `json:"on_sale"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

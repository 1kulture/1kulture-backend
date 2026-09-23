package requests

import "time"

type PromoCodeCreateRequest struct {
	Code string `json:"code" validate:"required,min=3,max=50,uppercase,alphanumeric"`
	Type string `json:"type" validate:"required,oneof=percentage fixed"`

	// For percentage: value in bps (500 = 5%). For fixed: value in minor units.
	ValueMinor int64 `json:"value_minor" validate:"required,min=1"`

	UsageLimit   int `json:"usage_limit" validate:"omitempty,min=0"`
	PerUserLimit int `json:"per_user_limit" validate:"omitempty,min=0"`

	ValidFrom *time.Time `json:"valid_from,omitempty"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`

	MinOrderMinor int64 `json:"min_order_minor" validate:"omitempty,min=0"`

	ApplicableTicketTypeIDs []string `json:"applicable_ticket_type_ids,omitempty" validate:"omitempty,max=20,dive,uuid"`

	IsActive *bool `json:"is_active,omitempty"`
}

type PromoCodeUpdateRequest struct {
	UsageLimit    *int       `json:"usage_limit,omitempty" validate:"omitempty,min=0"`
	PerUserLimit  *int       `json:"per_user_limit,omitempty" validate:"omitempty,min=0"`
	ValidFrom     *time.Time `json:"valid_from,omitempty"`
	ValidTo       *time.Time `json:"valid_to,omitempty"`
	MinOrderMinor *int64     `json:"min_order_minor,omitempty" validate:"omitempty,min=0"`
	IsActive      *bool      `json:"is_active,omitempty"`
}

type ValidatePromoRequest struct {
	EventID       string `json:"event_id" validate:"required,uuid"`
	PromoCode     string `json:"promo_code" validate:"required,min=3,max=50"`
	SubtotalMinor int64  `json:"subtotal_minor" validate:"required,min=0"`
}

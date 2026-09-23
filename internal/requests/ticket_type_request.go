package requests

import "time"

type TicketTypeCreateRequest struct {
	OccurrenceID *string `json:"occurrence_id,omitempty" validate:"omitempty,uuid"`

	Name        string `json:"name" validate:"required,min=2,max=255"`
	Description string `json:"description" validate:"omitempty"`
	ImageURL    string `json:"image_url" validate:"omitempty,url,max=500"`

	PriceMinor int64 `json:"price_minor" validate:"required,min=0"`

	QuantityTotal int `json:"quantity_total" validate:"required,min=1"`

	PerUserLimit int `json:"per_user_limit" validate:"omitempty,min=0"`

	SalesStartAt *time.Time `json:"sales_start_at,omitempty"`
	SalesEndAt   *time.Time `json:"sales_end_at,omitempty"`

	IsHidden    bool   `json:"is_hidden"`
	IsActive    *bool  `json:"is_active,omitempty"`
	AccessLevel string `json:"access_level" validate:"omitempty,oneof=general vip backstage"`

	RequiresApproval bool `json:"requires_approval"`
	SortOrder        int  `json:"sort_order" validate:"omitempty,min=0"`
}

type TicketTypeUpdateRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description *string `json:"description,omitempty"`
	ImageURL    *string `json:"image_url,omitempty" validate:"omitempty,url,max=500"`

	PriceMinor *int64 `json:"price_minor,omitempty" validate:"omitempty,min=0"`

	QuantityTotal *int `json:"quantity_total,omitempty" validate:"omitempty,min=1"`

	PerUserLimit *int `json:"per_user_limit,omitempty" validate:"omitempty,min=0"`

	SalesStartAt *time.Time `json:"sales_start_at,omitempty"`
	SalesEndAt   *time.Time `json:"sales_end_at,omitempty"`

	IsHidden    *bool   `json:"is_hidden,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
	AccessLevel *string `json:"access_level,omitempty" validate:"omitempty,oneof=general vip backstage"`

	RequiresApproval *bool `json:"requires_approval,omitempty"`
	SortOrder        *int  `json:"sort_order,omitempty" validate:"omitempty,min=0"`
}

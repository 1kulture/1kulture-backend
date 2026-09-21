package requests

// CreateCommissionTierRequest - Super Admin only.
type CreateCommissionTierRequest struct {
	Name            string `json:"name" validate:"required,min=2,max=100" example:"New Organizer"`
	MinRevenueMinor int64  `json:"min_revenue_minor" validate:"min=0" example:"0"`
	MaxRevenueMinor *int64 `json:"max_revenue_minor,omitempty" validate:"omitempty,min=0"`
	RateBps         int    `json:"rate_bps" validate:"required,min=0,max=10000" example:"700"` // 7%
	Priority        int    `json:"priority" validate:"omitempty,min=0"`
	IsActive        *bool  `json:"is_active,omitempty"`
}

// UpdateCommissionTierRequest - Super Admin only.
type UpdateCommissionTierRequest struct {
	Name            string `json:"name" validate:"omitempty,min=2,max=100"`
	MinRevenueMinor *int64 `json:"min_revenue_minor,omitempty" validate:"omitempty,min=0"`
	MaxRevenueMinor *int64 `json:"max_revenue_minor,omitempty" validate:"omitempty,min=0"`
	RateBps         *int   `json:"rate_bps,omitempty" validate:"omitempty,min=0,max=10000"`
	Priority        *int   `json:"priority,omitempty" validate:"omitempty,min=0"`
	IsActive        *bool  `json:"is_active,omitempty"`
}

// SetOrganizerCommissionOverrideRequest - Super Admin only.
// RateBps = -1 disables the override (falls back to tiered).
type SetOrganizerCommissionOverrideRequest struct {
	OrganizerID string `json:"organizer_id" validate:"required,uuid" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	RateBps     int    `json:"rate_bps" validate:"required,min=-1,max=10000" example:"500"`
	Reason      string `json:"reason" validate:"omitempty,max=500"`
}

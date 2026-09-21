package responses

import (
	"time"

	"github.com/google/uuid"
)

type CommissionTierResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	MinRevenueMinor int64     `json:"min_revenue_minor"`
	MaxRevenueMinor *int64    `json:"max_revenue_minor,omitempty"`
	RateBps         int       `json:"rate_bps"`
	RatePercentage  float64   `json:"rate_percentage"` // computed for convenience
	Priority        int       `json:"priority"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type OrganizerCommissionOverrideResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrganizerID    uuid.UUID  `json:"organizer_id"`
	RateBps        int        `json:"rate_bps"`
	RatePercentage float64    `json:"rate_percentage"`
	Reason         string     `json:"reason"`
	SetBy          *uuid.UUID `json:"set_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

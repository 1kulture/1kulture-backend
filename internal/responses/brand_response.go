package responses

import (
	"time"

	"github.com/google/uuid"
)

type BrandProfileResponse struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	BusinessName string    `json:"business_name"`
	LogoURL      string    `json:"logo_url,omitempty"`
	Industry     string    `json:"industry"`
	Description  string    `json:"description,omitempty"`
	Website      string    `json:"website,omitempty"`
	Instagram    string    `json:"instagram,omitempty"`
	Location     string    `json:"location"`
	ContactName  string    `json:"contact_name"`
	ContactEmail string    `json:"contact_email,omitempty"`
	ContactPhone string    `json:"contact_phone,omitempty"`
	BusinessSize string    `json:"business_size"`

	TargetAudience      string   `json:"target_audience,omitempty"`
	PreferredCategories []string `json:"preferred_categories,omitempty"`
	Tags                []string `json:"tags,omitempty"`

	IsActive bool `json:"is_active"`

	// Aggregated
	ActivePartnerships int `json:"active_partnerships,omitempty"`
	TotalPartnerships  int `json:"total_partnerships,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BrandSummary struct {
	ID           uuid.UUID `json:"id"`
	BusinessName string    `json:"business_name"`
	LogoURL      string    `json:"logo_url,omitempty"`
	Industry     string    `json:"industry"`
	Location     string    `json:"location"`
}

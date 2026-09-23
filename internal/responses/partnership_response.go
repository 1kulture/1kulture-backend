package responses

import (
	"time"

	"github.com/google/uuid"
)

type EventPartnershipConfigResponse struct {
	ID      uuid.UUID `json:"id"`
	EventID uuid.UUID `json:"event_id"`
	IsOpen  bool      `json:"is_open"`

	PartnershipTypes []string `json:"partnership_types,omitempty"`
	BrandBenefits    []string `json:"brand_benefits,omitempty"`

	ExpectedAttendees int      `json:"expected_attendees"`
	AudienceAgeRange  string   `json:"audience_age_range,omitempty"`
	AudienceLocations []string `json:"audience_locations,omitempty"`
	AudienceInterests []string `json:"audience_interests,omitempty"`

	Notes string `json:"notes,omitempty"`

	Opportunities []OpportunityResponse `json:"opportunities,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OpportunityResponse struct {
	ID          uuid.UUID `json:"id"`
	EventID     uuid.UUID `json:"event_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Type        string    `json:"type"`

	BudgetMinMinor int64 `json:"budget_min_minor"`
	BudgetMaxMinor int64 `json:"budget_max_minor"`

	SlotsTotal     int `json:"slots_total"`
	SlotsRemaining int `json:"slots_remaining"`

	SortOrder int  `json:"sort_order"`
	IsActive  bool `json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PartnershipRequestResponse struct {
	ID             uuid.UUID  `json:"id"`
	BrandProfileID uuid.UUID  `json:"brand_profile_id"`
	BrandUserID    uuid.UUID  `json:"brand_user_id"`
	EventID        uuid.UUID  `json:"event_id"`
	OrganizerID    uuid.UUID  `json:"organizer_id"`
	OpportunityID  *uuid.UUID `json:"opportunity_id,omitempty"`

	RequestedTypes []string `json:"requested_types"`
	Message        string   `json:"message,omitempty"`

	Status string `json:"status"`

	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
	ActivatedAt *time.Time `json:"activated_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DeclinedAt  *time.Time `json:"declined_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	Reason string `json:"reason,omitempty"`

	// Embedded summary info for cards
	Brand *BrandSummary        `json:"brand,omitempty"`
	Event *PartnershipEventRef `json:"event,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PartnershipEventRef is a compact event reference used in partnership views.
type PartnershipEventRef struct {
	ID           uuid.UUID `json:"id"`
	Slug         string    `json:"slug"`
	Title        string    `json:"title"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	StartAt      time.Time `json:"start_at"`
	EndAt        time.Time `json:"end_at"`
	Currency     string    `json:"currency"`
}

type PartnershipMetricResponse struct {
	ID            uuid.UUID `json:"id"`
	PartnershipID uuid.UUID `json:"partnership_id"`
	Key           string    `json:"key"`
	Value         int64     `json:"value"`
	Currency      string    `json:"currency,omitempty"`
	Note          string    `json:"note,omitempty"`
	ReportedByID  uuid.UUID `json:"reported_by_id"`
	ReportedRole  string    `json:"reported_role"`
	CreatedAt     time.Time `json:"created_at"`
}

type PartnershipDashboardResponse struct {
	PendingCount   int `json:"pending_count"`
	AcceptedCount  int `json:"accepted_count"`
	ActiveCount    int `json:"active_count"`
	CompletedCount int `json:"completed_count"`
	DeclinedCount  int `json:"declined_count"`
	CancelledCount int `json:"cancelled_count"`
	TotalCount     int `json:"total_count"`
}

type BrandDashboardResponse struct {
	ActivePartnerships int `json:"active_partnerships"`
	PendingRequests    int `json:"pending_requests"`
	AvailableEvents    int `json:"available_events"`
}

type AffiliateCodeResponse struct {
	ID              uuid.UUID `json:"id"`
	PartnershipID   uuid.UUID `json:"partnership_id"`
	PromoCodeID     uuid.UUID `json:"promo_code_id"`
	Code            string    `json:"code"`
	Clicks          int64     `json:"clicks"`
	Conversions     int64     `json:"conversions"`
	RevenueMinor    int64     `json:"revenue_minor"`
	CommissionMinor int64     `json:"commission_minor"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

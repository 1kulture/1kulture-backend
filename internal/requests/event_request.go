package requests

import "time"

// EventCreateRequest creates a new draft event.
type EventCreateRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=255" example:"Lagos Tech Summit 2026"`
	Summary     string `json:"summary" validate:"omitempty,max=500" example:"The biggest tech gathering in West Africa."`
	Description string `json:"description" validate:"omitempty" example:"Full markdown description..."`

	BannerURL    string   `json:"banner_url" validate:"omitempty,url,max=500"`
	ThumbnailURL string   `json:"thumbnail_url" validate:"omitempty,url,max=500"`
	Gallery      []string `json:"gallery" validate:"omitempty,max=10,dive,url"`

	Category string   `json:"category" validate:"required,min=2,max=100" example:"technology"`
	Tags     []string `json:"tags" validate:"omitempty,max=20,dive,min=1,max=50"`

	EventType string `json:"event_type" validate:"required,oneof=in_person virtual hybrid" example:"in_person"`

	// Venue fields (required for in_person/hybrid)
	VenueName       string   `json:"venue_name" validate:"omitempty,max=255"`
	VenueAddress    string   `json:"venue_address" validate:"omitempty,max=500"`
	VenueCity       string   `json:"venue_city" validate:"omitempty,max=100"`
	VenueState      string   `json:"venue_state" validate:"omitempty,max=100"`
	VenueCountry    string   `json:"venue_country" validate:"omitempty,max=100"`
	VenuePostalCode string   `json:"venue_postal_code" validate:"omitempty,max=20"`
	VenueLat        *float64 `json:"venue_lat" validate:"omitempty"`
	VenueLng        *float64 `json:"venue_lng" validate:"omitempty"`

	// Virtual fields (required for virtual/hybrid)
	VirtualURL      string `json:"virtual_url" validate:"omitempty,url,max=500"`
	VirtualPlatform string `json:"virtual_platform" validate:"omitempty,max=100"`

	Timezone    string     `json:"timezone" validate:"required" example:"Africa/Lagos"`
	StartAt     time.Time  `json:"start_at" validate:"required"`
	EndAt       time.Time  `json:"end_at" validate:"required"`
	DoorsOpenAt *time.Time `json:"doors_open_at,omitempty"`

	Capacity       *int   `json:"capacity" validate:"omitempty,min=1"`
	AgeRestriction string `json:"age_restriction" validate:"omitempty,oneof=all_ages 13_plus 16_plus 18_plus 21_plus"`

	Visibility string `json:"visibility" validate:"omitempty,oneof=public unlisted private"`

	RefundPolicy     string `json:"refund_policy" validate:"omitempty"`
	RefundPolicyDays int    `json:"refund_policy_days" validate:"omitempty,min=0,max=365"`

	Currency string `json:"currency" validate:"omitempty,len=3,uppercase" example:"NGN"`
}

// EventUpdateRequest allows partial updates.
type EventUpdateRequest struct {
	Title        *string  `json:"title,omitempty" validate:"omitempty,min=3,max=255"`
	Summary      *string  `json:"summary,omitempty" validate:"omitempty,max=500"`
	Description  *string  `json:"description,omitempty"`
	BannerURL    *string  `json:"banner_url,omitempty" validate:"omitempty,url,max=500"`
	ThumbnailURL *string  `json:"thumbnail_url,omitempty" validate:"omitempty,url,max=500"`
	Gallery      []string `json:"gallery,omitempty" validate:"omitempty,max=10,dive,url"`
	Category     *string  `json:"category,omitempty" validate:"omitempty,min=2,max=100"`
	Tags         []string `json:"tags,omitempty" validate:"omitempty,max=20,dive,min=1,max=50"`

	EventType *string `json:"event_type,omitempty" validate:"omitempty,oneof=in_person virtual hybrid"`

	VenueName       *string  `json:"venue_name,omitempty"`
	VenueAddress    *string  `json:"venue_address,omitempty"`
	VenueCity       *string  `json:"venue_city,omitempty"`
	VenueState      *string  `json:"venue_state,omitempty"`
	VenueCountry    *string  `json:"venue_country,omitempty"`
	VenuePostalCode *string  `json:"venue_postal_code,omitempty"`
	VenueLat        *float64 `json:"venue_lat,omitempty"`
	VenueLng        *float64 `json:"venue_lng,omitempty"`

	VirtualURL      *string `json:"virtual_url,omitempty" validate:"omitempty,url,max=500"`
	VirtualPlatform *string `json:"virtual_platform,omitempty"`

	Timezone    *string    `json:"timezone,omitempty"`
	StartAt     *time.Time `json:"start_at,omitempty"`
	EndAt       *time.Time `json:"end_at,omitempty"`
	DoorsOpenAt *time.Time `json:"doors_open_at,omitempty"`

	Capacity       *int    `json:"capacity,omitempty" validate:"omitempty,min=1"`
	AgeRestriction *string `json:"age_restriction,omitempty" validate:"omitempty,oneof=all_ages 13_plus 16_plus 18_plus 21_plus"`

	Visibility *string `json:"visibility,omitempty" validate:"omitempty,oneof=public unlisted private"`

	RefundPolicy     *string `json:"refund_policy,omitempty"`
	RefundPolicyDays *int    `json:"refund_policy_days,omitempty" validate:"omitempty,min=0,max=365"`

	Currency *string `json:"currency,omitempty" validate:"omitempty,len=3,uppercase"`
}

// EventCancelRequest cancels an event.
type EventCancelRequest struct {
	Reason string `json:"reason" validate:"required,min=5,max=500" example:"Unforeseen venue closure."`
}

// EventPostponeRequest postpones an event.
type EventPostponeRequest struct {
	NewStartAt time.Time `json:"new_start_at" validate:"required"`
	NewEndAt   time.Time `json:"new_end_at" validate:"required"`
	Reason     string    `json:"reason" validate:"required,min=5,max=500"`
}

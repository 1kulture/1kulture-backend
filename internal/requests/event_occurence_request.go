package requests

import "time"

type OccurrenceCreateRequest struct {
	SessionName  string    `json:"session_name" validate:"required,min=2,max=255"`
	StartAt      time.Time `json:"start_at" validate:"required"`
	EndAt        time.Time `json:"end_at" validate:"required"`
	VenueName    string    `json:"venue_name" validate:"omitempty,max=255"`
	VenueAddress string    `json:"venue_address" validate:"omitempty,max=500"`
	VirtualURL   string    `json:"virtual_url" validate:"omitempty,url,max=500"`
	Capacity     *int      `json:"capacity" validate:"omitempty,min=1"`
	SortOrder    int       `json:"sort_order" validate:"omitempty,min=0"`
}

type OccurrenceUpdateRequest struct {
	SessionName  *string    `json:"session_name,omitempty" validate:"omitempty,min=2,max=255"`
	StartAt      *time.Time `json:"start_at,omitempty"`
	EndAt        *time.Time `json:"end_at,omitempty"`
	VenueName    *string    `json:"venue_name,omitempty"`
	VenueAddress *string    `json:"venue_address,omitempty"`
	VirtualURL   *string    `json:"virtual_url,omitempty" validate:"omitempty,url,max=500"`
	Capacity     *int       `json:"capacity,omitempty" validate:"omitempty,min=1"`
	SortOrder    *int       `json:"sort_order,omitempty" validate:"omitempty,min=0"`
}

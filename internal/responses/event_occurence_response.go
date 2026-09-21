package responses

import (
	"time"

	"github.com/google/uuid"
)

type EventOccurrenceResponse struct {
	ID           uuid.UUID `json:"id"`
	EventID      uuid.UUID `json:"event_id"`
	SessionName  string    `json:"session_name"`
	StartAt      time.Time `json:"start_at"`
	EndAt        time.Time `json:"end_at"`
	VenueName    string    `json:"venue_name,omitempty"`
	VenueAddress string    `json:"venue_address,omitempty"`
	VirtualURL   string    `json:"virtual_url,omitempty"`
	Capacity     *int      `json:"capacity,omitempty"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

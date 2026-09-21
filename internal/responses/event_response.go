package responses

import (
	"time"

	"github.com/google/uuid"
)

type EventResponse struct {
	ID          uuid.UUID `json:"id"`
	OrganizerID uuid.UUID `json:"organizer_id"`
	Slug        string    `json:"slug"`

	Title        string   `json:"title"`
	Summary      string   `json:"summary,omitempty"`
	Description  string   `json:"description,omitempty"`
	BannerURL    string   `json:"banner_url,omitempty"`
	ThumbnailURL string   `json:"thumbnail_url,omitempty"`
	Gallery      []string `json:"gallery,omitempty"`

	Category string   `json:"category"`
	Tags     []string `json:"tags,omitempty"`

	EventType string `json:"event_type"`

	VenueName       string   `json:"venue_name,omitempty"`
	VenueAddress    string   `json:"venue_address,omitempty"`
	VenueCity       string   `json:"venue_city,omitempty"`
	VenueState      string   `json:"venue_state,omitempty"`
	VenueCountry    string   `json:"venue_country,omitempty"`
	VenuePostalCode string   `json:"venue_postal_code,omitempty"`
	VenueLat        *float64 `json:"venue_lat,omitempty"`
	VenueLng        *float64 `json:"venue_lng,omitempty"`

	VirtualURL      string `json:"virtual_url,omitempty"`
	VirtualPlatform string `json:"virtual_platform,omitempty"`

	Timezone    string     `json:"timezone"`
	StartAt     time.Time  `json:"start_at"`
	EndAt       time.Time  `json:"end_at"`
	DoorsOpenAt *time.Time `json:"doors_open_at,omitempty"`

	Capacity       *int   `json:"capacity,omitempty"`
	AgeRestriction string `json:"age_restriction"`

	Status     string `json:"status"`
	Visibility string `json:"visibility"`
	IsFeatured bool   `json:"is_featured"`

	RefundPolicy     string `json:"refund_policy,omitempty"`
	RefundPolicyDays int    `json:"refund_policy_days"`

	Currency string `json:"currency"`

	PublishedAt  *time.Time `json:"published_at,omitempty"`
	CancelledAt  *time.Time `json:"cancelled_at,omitempty"`
	CancelReason string     `json:"cancel_reason,omitempty"`

	FollowerCount int `json:"follower_count"`
	ShareCount    int `json:"share_count"`

	Occurrences  []EventOccurrenceResponse  `json:"occurrences,omitempty"`
	CoOrganizers []EventCoOrganizerResponse `json:"co_organizers,omitempty"`

	Organizer *EventOrganizerBrief `json:"organizer,omitempty"`

	IsFollowing bool `json:"is_following"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EventSummaryResponse is the lighter payload for lists.
type EventSummaryResponse struct {
	ID           uuid.UUID `json:"id"`
	Slug         string    `json:"slug"`
	Title        string    `json:"title"`
	Summary      string    `json:"summary,omitempty"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	BannerURL    string    `json:"banner_url,omitempty"`
	Category     string    `json:"category"`
	EventType    string    `json:"event_type"`
	VenueCity    string    `json:"venue_city,omitempty"`
	VenueCountry string    `json:"venue_country,omitempty"`
	Timezone     string    `json:"timezone"`
	StartAt      time.Time `json:"start_at"`
	EndAt        time.Time `json:"end_at"`
	Status       string    `json:"status"`
	Visibility   string    `json:"visibility"`
	Currency     string    `json:"currency"`
	IsFeatured   bool      `json:"is_featured"`
	OrganizerID  uuid.UUID `json:"organizer_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type EventOrganizerBrief struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
}

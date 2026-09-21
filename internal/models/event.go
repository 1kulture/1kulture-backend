package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusPublished EventStatus = "published"
	EventStatusPostponed EventStatus = "postponed"
	EventStatusCancelled EventStatus = "cancelled"
	EventStatusCompleted EventStatus = "completed"
	EventStatusArchived  EventStatus = "archived"
)

type EventVisibility string

const (
	EventVisibilityPublic   EventVisibility = "public"
	EventVisibilityUnlisted EventVisibility = "unlisted"
	EventVisibilityPrivate  EventVisibility = "private"
)

type EventType string

const (
	EventTypeInPerson EventType = "in_person"
	EventTypeVirtual  EventType = "virtual"
	EventTypeHybrid   EventType = "hybrid"
)

type AgeRestriction string

const (
	AgeAllAges AgeRestriction = "all_ages"
	Age13Plus  AgeRestriction = "13_plus"
	Age16Plus  AgeRestriction = "16_plus"
	Age18Plus  AgeRestriction = "18_plus"
	Age21Plus  AgeRestriction = "21_plus"
)

type Event struct {
	BaseModel

	OrganizerID uuid.UUID `gorm:"type:uuid;not null;index" json:"organizer_id"`
	Slug        string    `gorm:"uniqueIndex;not null;size:255" json:"slug"`

	Title        string         `gorm:"not null;size:255" json:"title"`
	Summary      string         `gorm:"size:500" json:"summary"`
	Description  string         `gorm:"type:text" json:"description"`
	BannerURL    string         `gorm:"size:500" json:"banner_url"`
	ThumbnailURL string         `gorm:"size:500" json:"thumbnail_url"`
	Gallery      datatypes.JSON `gorm:"type:jsonb" json:"gallery"` // []string of URLs

	Category string         `gorm:"size:100;index" json:"category"`
	Tags     datatypes.JSON `gorm:"type:jsonb" json:"tags"` // []string

	EventType EventType `gorm:"size:20;not null;default:'in_person';index" json:"event_type"`

	// Venue (in-person / hybrid)
	VenueName       string   `gorm:"size:255" json:"venue_name"`
	VenueAddress    string   `gorm:"size:500" json:"venue_address"`
	VenueCity       string   `gorm:"size:100;index" json:"venue_city"`
	VenueState      string   `gorm:"size:100" json:"venue_state"`
	VenueCountry    string   `gorm:"size:100;index" json:"venue_country"`
	VenuePostalCode string   `gorm:"size:20" json:"venue_postal_code"`
	VenueLat        *float64 `json:"venue_lat,omitempty"`
	VenueLng        *float64 `json:"venue_lng,omitempty"`
	VenuePlaceID    string   `gorm:"size:255" json:"venue_place_id"`

	// Virtual (virtual / hybrid)
	VirtualURL      string `gorm:"size:500" json:"virtual_url"`
	VirtualPlatform string `gorm:"size:100" json:"virtual_platform"`

	Timezone    string     `gorm:"size:64;not null;default:'UTC'" json:"timezone"`
	StartAt     time.Time  `gorm:"not null;index" json:"start_at"`
	EndAt       time.Time  `gorm:"not null;index" json:"end_at"`
	DoorsOpenAt *time.Time `json:"doors_open_at,omitempty"`

	Capacity       *int           `json:"capacity,omitempty"`
	AgeRestriction AgeRestriction `gorm:"size:20;default:'all_ages'" json:"age_restriction"`

	Status     EventStatus     `gorm:"size:20;not null;default:'draft';index" json:"status"`
	Visibility EventVisibility `gorm:"size:20;not null;default:'public';index" json:"visibility"`
	IsFeatured bool            `gorm:"default:false;index" json:"is_featured"`

	RefundPolicy     string `gorm:"type:text" json:"refund_policy"`
	RefundPolicyDays int    `gorm:"default:7" json:"refund_policy_days"`

	Currency string         `gorm:"size:3;not null;default:'NGN'" json:"currency"`
	Meta     datatypes.JSON `gorm:"type:jsonb" json:"meta"`

	PublishedAt  *time.Time `json:"published_at,omitempty"`
	CancelledAt  *time.Time `json:"cancelled_at,omitempty"`
	CancelReason string     `gorm:"size:500" json:"cancel_reason"`

	// Relationships
	Organizer    User               `gorm:"foreignKey:OrganizerID" json:"-"`
	Occurrences  []EventOccurrence  `gorm:"foreignKey:EventID" json:"occurrences,omitempty"`
	CoOrganizers []EventCoOrganizer `gorm:"foreignKey:EventID" json:"co_organizers,omitempty"`
	Staff        []EventStaff       `gorm:"foreignKey:EventID" json:"staff,omitempty"`
}

func (e *Event) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.Status == "" {
		e.Status = EventStatusDraft
	}
	if e.Visibility == "" {
		e.Visibility = EventVisibilityPublic
	}
	if e.EventType == "" {
		e.EventType = EventTypeInPerson
	}
	if e.Timezone == "" {
		e.Timezone = "UTC"
	}
	if e.Currency == "" {
		e.Currency = "NGN"
	}
	if e.AgeRestriction == "" {
		e.AgeRestriction = AgeAllAges
	}
	return nil
}

func (e *Event) IsPublished() bool {
	return e.Status == EventStatusPublished
}

func (e *Event) IsCancelled() bool {
	return e.Status == EventStatusCancelled
}

package responses

import (
	"time"

	"github.com/google/uuid"
)

type EventCoOrganizerResponse struct {
	ID          uuid.UUID              `json:"id"`
	EventID     uuid.UUID              `json:"event_id"`
	UserID      uuid.UUID              `json:"user_id"`
	Role        string                 `json:"role"`
	Permissions map[string]interface{} `json:"permissions,omitempty"`
	FirstName   string                 `json:"first_name,omitempty"`
	LastName    string                 `json:"last_name,omitempty"`
	Email       string                 `json:"email,omitempty"`
	AvatarURL   string                 `json:"avatar_url,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

type EventStaffResponse struct {
	ID        uuid.UUID `json:"id"`
	EventID   uuid.UUID `json:"event_id"`
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type EventFollowerResponse struct {
	UserID            uuid.UUID `json:"user_id"`
	FirstName         string    `json:"first_name,omitempty"`
	LastName          string    `json:"last_name,omitempty"`
	AvatarURL         string    `json:"avatar_url,omitempty"`
	NotifyNewSessions bool      `json:"notify_new_sessions"`
	NotifyUpdates     bool      `json:"notify_updates"`
	FollowedAt        time.Time `json:"followed_at"`
}

type OrganizerFollowerResponse struct {
	UserID          uuid.UUID `json:"user_id"`
	OrganizerID     uuid.UUID `json:"organizer_id"`
	NotifyNewEvents bool      `json:"notify_new_events"`
	FollowedAt      time.Time `json:"followed_at"`
}

type EventShareResponse struct {
	ID           uuid.UUID `json:"id"`
	EventID      uuid.UUID `json:"event_id"`
	Channel      string    `json:"channel"`
	ReferralCode string    `json:"referral_code,omitempty"`
	SharedAt     time.Time `json:"shared_at"`
}

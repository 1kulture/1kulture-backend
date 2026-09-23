package responses

import (
	"time"

	"github.com/google/uuid"
)

type NotificationResponse struct {
	ID      uuid.UUID `json:"id"`
	UserID  uuid.UUID `json:"user_id"`
	Type    string    `json:"type"`
	Channel string    `json:"channel"`

	Title string `json:"title"`
	Body  string `json:"body"`

	RefType string    `json:"ref_type,omitempty"`
	RefID   uuid.UUID `json:"ref_id,omitempty"`

	Data   map[string]interface{} `json:"data,omitempty"`
	ReadAt *time.Time             `json:"read_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

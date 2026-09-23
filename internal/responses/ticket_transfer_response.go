package responses

import (
	"time"

	"github.com/google/uuid"
)

type TicketTransferResponse struct {
	ID         uuid.UUID  `json:"id"`
	TicketID   uuid.UUID  `json:"ticket_id"`
	FromUserID uuid.UUID  `json:"from_user_id"`
	ToEmail    string     `json:"to_email"`
	ToUserID   *uuid.UUID `json:"to_user_id,omitempty"`

	Message string `json:"message,omitempty"`

	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`

	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
	DeclinedAt  *time.Time `json:"declined_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

package responses

import (
	"time"

	"github.com/google/uuid"
)

type CheckInRecordResponse struct {
	ID        uuid.UUID `json:"id"`
	TicketID  uuid.UUID `json:"ticket_id"`
	EventID   uuid.UUID `json:"event_id"`
	ScannedBy uuid.UUID `json:"scanned_by"`
	Method    string    `json:"method"`
	ScannedAt time.Time `json:"scanned_at"`
	Location  string    `json:"location,omitempty"`
}

type CheckInResponse struct {
	Ticket         TicketResponse `json:"ticket"`
	CheckedInAt    time.Time      `json:"checked_in_at"`
	ScannedBy      uuid.UUID      `json:"scanned_by"`
	AlreadyScanned bool           `json:"already_scanned"`
}

type CheckInStatsResponse struct {
	EventID          uuid.UUID `json:"event_id"`
	TotalTickets     int       `json:"total_tickets"`
	CheckedIn        int       `json:"checked_in"`
	Remaining        int       `json:"remaining"`
	CheckedInPercent float64   `json:"checked_in_percent"`
}

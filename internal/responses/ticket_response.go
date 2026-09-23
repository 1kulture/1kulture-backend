package responses

import (
	"time"

	"github.com/google/uuid"
)

type TicketResponse struct {
	ID           uuid.UUID  `json:"id"`
	OrderID      uuid.UUID  `json:"order_id"`
	OrderItemID  uuid.UUID  `json:"order_item_id"`
	TicketTypeID uuid.UUID  `json:"ticket_type_id"`
	EventID      uuid.UUID  `json:"event_id"`
	OccurrenceID *uuid.UUID `json:"occurrence_id,omitempty"`
	UserID       uuid.UUID  `json:"user_id"`

	Code string `json:"code"`

	HolderName  string `json:"holder_name"`
	HolderEmail string `json:"holder_email,omitempty"`
	HolderPhone string `json:"holder_phone,omitempty"`

	Status string `json:"status"`

	IssuedAt     time.Time  `json:"issued_at"`
	UsedAt       *time.Time `json:"used_at,omitempty"`
	ScanLocation string     `json:"scan_location,omitempty"`

	TransferCount int `json:"transfer_count"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TicketWithQRResponse is returned at issuance so the user can download the QR.
type TicketWithQRResponse struct {
	TicketResponse
	QRPayload string `json:"qr_payload"`
}

type TicketCheckInResponse struct {
	Ticket      TicketResponse `json:"ticket"`
	CheckedInAt time.Time      `json:"checked_in_at"`
	ScannedBy   uuid.UUID      `json:"scanned_by"`
}

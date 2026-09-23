package responses

import (
	"time"

	"github.com/google/uuid"
)

type OrderItemResponse struct {
	ID             uuid.UUID `json:"id"`
	TicketTypeID   uuid.UUID `json:"ticket_type_id"`
	TicketTypeName string    `json:"ticket_type_name"`
	Quantity       int       `json:"quantity"`
	UnitPriceMinor int64     `json:"unit_price_minor"`
	TotalMinor     int64     `json:"total_minor"`
}

type OrderResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	EventID   uuid.UUID `json:"event_id"`
	Reference string    `json:"reference"`

	Status string `json:"status"`

	Currency           string `json:"currency"`
	SubtotalMinor      int64  `json:"subtotal_minor"`
	DiscountMinor      int64  `json:"discount_minor"`
	PlatformFeeMinor   int64  `json:"platform_fee_minor"`
	ProcessingFeeMinor int64  `json:"processing_fee_minor"`
	TotalMinor         int64  `json:"total_minor"`

	CommissionRateBps int   `json:"commission_rate_bps"`
	CommissionMinor   int64 `json:"commission_minor"`
	OrganizerNetMinor int64 `json:"organizer_net_minor"`

	PaymentProvider  string     `json:"payment_provider,omitempty"`
	PaymentReference string     `json:"payment_reference,omitempty"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`

	EscrowStatus     string     `json:"escrow_status"`
	EscrowReleaseAt  *time.Time `json:"escrow_release_at,omitempty"`
	EscrowReleasedAt *time.Time `json:"escrow_released_at,omitempty"`

	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	Items   []OrderItemResponse `json:"items,omitempty"`
	Tickets []TicketResponse    `json:"tickets,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type InitializePaymentResponse struct {
	OrderID          uuid.UUID `json:"order_id"`
	Reference        string    `json:"reference"`
	AuthorizationURL string    `json:"authorization_url"`
	AccessCode       string    `json:"access_code,omitempty"`
	ExpiresAt        time.Time `json:"expires_at"`
}

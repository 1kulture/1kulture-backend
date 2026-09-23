package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending           OrderStatus = "pending"
	OrderStatusPaid              OrderStatus = "paid"
	OrderStatusFailed            OrderStatus = "failed"
	OrderStatusCancelled         OrderStatus = "cancelled"
	OrderStatusRefunded          OrderStatus = "refunded"
	OrderStatusPartiallyRefunded OrderStatus = "partially_refunded"
	OrderStatusExpired           OrderStatus = "expired"
)

type EscrowStatus string

const (
	EscrowStatusNone              EscrowStatus = "none"    // no escrow (e.g. free tickets)
	EscrowStatusPending           EscrowStatus = "pending" // paid, funds held
	EscrowStatusHeld              EscrowStatus = "held"
	EscrowStatusReleased          EscrowStatus = "released"
	EscrowStatusRefunded          EscrowStatus = "refunded"
	EscrowStatusPartiallyRefunded EscrowStatus = "partially_refunded"
)

type Order struct {
	BaseModel

	UserID  uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	EventID uuid.UUID `gorm:"type:uuid;not null;index" json:"event_id"`

	Reference string      `gorm:"uniqueIndex;not null;size:64" json:"reference"` // internal ref
	Status    OrderStatus `gorm:"size:30;not null;default:'pending';index" json:"status"`

	// Pricing
	Currency           string `gorm:"size:3;not null" json:"currency"`
	SubtotalMinor      int64  `gorm:"not null;default:0" json:"subtotal_minor"`
	DiscountMinor      int64  `gorm:"not null;default:0" json:"discount_minor"`
	PlatformFeeMinor   int64  `gorm:"not null;default:0" json:"platform_fee_minor"`
	ProcessingFeeMinor int64  `gorm:"not null;default:0" json:"processing_fee_minor"`
	TotalMinor         int64  `gorm:"not null;default:0" json:"total_minor"`

	// Commission snapshot (bps at time of purchase)
	CommissionRateBps int   `gorm:"not null;default:0" json:"commission_rate_bps"`
	CommissionMinor   int64 `gorm:"not null;default:0" json:"commission_minor"`
	OrganizerNetMinor int64 `gorm:"not null;default:0" json:"organizer_net_minor"`

	// Promo
	PromoCodeID *uuid.UUID `gorm:"type:uuid" json:"promo_code_id,omitempty"`

	// Payment provider data
	PaymentProvider  string     `gorm:"size:30" json:"payment_provider,omitempty"`
	PaymentReference string     `gorm:"size:100;index" json:"payment_reference,omitempty"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`

	// Escrow
	EscrowStatus     EscrowStatus `gorm:"size:30;not null;default:'none';index" json:"escrow_status"`
	EscrowReleaseAt  *time.Time   `json:"escrow_release_at,omitempty"`
	EscrowReleasedAt *time.Time   `json:"escrow_released_at,omitempty"`

	// Idempotency
	IdempotencyKey string `gorm:"size:100;index" json:"-"`

	// Buyer details (snapshot — may differ from user profile)
	BillingInfo datatypes.JSON `gorm:"type:jsonb" json:"billing_info,omitempty"`

	// Reservation expiry for pending orders
	ExpiresAt *time.Time `gorm:"index" json:"expires_at,omitempty"`

	// Relationships
	User    User        `gorm:"foreignKey:UserID" json:"-"`
	Event   Event       `gorm:"foreignKey:EventID" json:"-"`
	Items   []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Tickets []Ticket    `gorm:"foreignKey:OrderID" json:"tickets,omitempty"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	if o.Status == "" {
		o.Status = OrderStatusPending
	}
	if o.EscrowStatus == "" {
		o.EscrowStatus = EscrowStatusNone
	}
	return nil
}

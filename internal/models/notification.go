package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type NotificationChannel string

const (
	NotificationChannelInApp NotificationChannel = "in_app"
	NotificationChannelEmail NotificationChannel = "email"
	NotificationChannelPush  NotificationChannel = "push"
	NotificationChannelSMS   NotificationChannel = "sms"
)

type NotificationType string

const (
	// Brand partnership notifications
	NotificationPartnershipRequested   NotificationType = "partnership_requested"
	NotificationPartnershipAccepted    NotificationType = "partnership_accepted"
	NotificationPartnershipDeclined    NotificationType = "partnership_declined"
	NotificationPartnershipActivated   NotificationType = "partnership_activated"
	NotificationPartnershipCompleted   NotificationType = "partnership_completed"
	NotificationPartnershipCancelled   NotificationType = "partnership_cancelled"
	NotificationPartnershipMetricAdded NotificationType = "partnership_metric_added"

	// Ticketing notifications
	NotificationOrderPaid         NotificationType = "order_paid"
	NotificationTicketTransferred NotificationType = "ticket_transferred"
	NotificationTicketReceived    NotificationType = "ticket_received"
	NotificationRefundProcessed   NotificationType = "refund_processed"

	// Event notifications
	NotificationEventPublished NotificationType = "event_published"
	NotificationEventCancelled NotificationType = "event_cancelled"
)

type Notification struct {
	BaseModel

	UserID  uuid.UUID           `gorm:"type:uuid;not null;index" json:"user_id"`
	Type    NotificationType    `gorm:"size:50;not null;index" json:"type"`
	Channel NotificationChannel `gorm:"size:20;not null;default:'in_app'" json:"channel"`

	Title string `gorm:"size:255;not null" json:"title"`
	Body  string `gorm:"type:text" json:"body"`

	// Polymorphic reference (e.g. "partnership", <id>)
	RefType string    `gorm:"size:50;index" json:"ref_type,omitempty"`
	RefID   uuid.UUID `gorm:"type:uuid;index" json:"ref_id,omitempty"`

	// Extra data for the client (icons, action urls, etc.)
	Data datatypes.JSON `gorm:"type:jsonb" json:"data,omitempty"`

	ReadAt     *time.Time `gorm:"index" json:"read_at,omitempty"`
	SentAt     *time.Time `json:"sent_at,omitempty"`
	FailedAt   *time.Time `json:"failed_at,omitempty"`
	FailReason string     `gorm:"size:500" json:"fail_reason,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

func (n *Notification) IsRead() bool {
	return n.ReadAt != nil
}

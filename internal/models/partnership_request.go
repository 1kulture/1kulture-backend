package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PartnershipRequestStatus string

const (
	PartnershipStatusPending   PartnershipRequestStatus = "pending"
	PartnershipStatusAccepted  PartnershipRequestStatus = "accepted"
	PartnershipStatusActive    PartnershipRequestStatus = "active"
	PartnershipStatusCompleted PartnershipRequestStatus = "completed"
	PartnershipStatusDeclined  PartnershipRequestStatus = "declined"
	PartnershipStatusCancelled PartnershipRequestStatus = "cancelled"
)

// PartnershipRequest is a brand's request to partner with an event.
type PartnershipRequest struct {
	BaseModel

	BrandProfileID uuid.UUID `gorm:"type:uuid;not null;index" json:"brand_profile_id"`
	BrandUserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"brand_user_id"`
	EventID        uuid.UUID `gorm:"type:uuid;not null;index" json:"event_id"`
	OrganizerID    uuid.UUID `gorm:"type:uuid;not null;index" json:"organizer_id"`

	// The specific opportunity they're interested in (optional — brands may have a general request)
	OpportunityID *uuid.UUID `gorm:"type:uuid;index" json:"opportunity_id,omitempty"`

	// What the brand wants
	RequestedTypes datatypes.JSON `gorm:"type:jsonb" json:"requested_types"` // []PartnershipType
	Message        string         `gorm:"type:text" json:"message"`

	Status PartnershipRequestStatus `gorm:"size:20;not null;default:'pending';index" json:"status"`

	// Timestamps for state transitions
	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
	ActivatedAt *time.Time `json:"activated_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	DeclinedAt  *time.Time `json:"declined_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	// Organizer's reason for decline/cancel (optional)
	Reason string `gorm:"size:500" json:"reason,omitempty"`

	BrandProfile BrandProfile `gorm:"foreignKey:BrandProfileID" json:"-"`
	Event        Event        `gorm:"foreignKey:EventID" json:"-"`
}

func (p *PartnershipRequest) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.Status == "" {
		p.Status = PartnershipStatusPending
	}
	return nil
}

// IsActiveState returns true if the partnership is in accepted/active.
func (p *PartnershipRequest) IsActiveState() bool {
	return p.Status == PartnershipStatusAccepted || p.Status == PartnershipStatusActive
}

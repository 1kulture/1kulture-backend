package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// PartnershipMetric is a reported metric for a partnership.
// Example: reach, impressions, clicks, promo_code_uses, estimated_sales, tickets_sold.
type PartnershipMetric struct {
	BaseModel
	PartnershipID uuid.UUID `gorm:"type:uuid;not null;index" json:"partnership_id"`

	Key   string `gorm:"size:50;not null;index" json:"key"`
	Value int64  `gorm:"not null;default:0" json:"value"`

	// Optional currency for monetary metrics (estimated_sales)
	Currency string `gorm:"size:3" json:"currency,omitempty"`

	// Free-form context
	Note string `gorm:"type:text" json:"note,omitempty"`

	// Who reported it (organizer or brand)
	ReportedByID uuid.UUID      `gorm:"type:uuid;not null" json:"reported_by_id"`
	ReportedRole string         `gorm:"size:20;not null" json:"reported_role"` // organizer | brand
	Meta         datatypes.JSON `gorm:"type:jsonb" json:"meta,omitempty"`

	Partnership PartnershipRequest `gorm:"foreignKey:PartnershipID" json:"-"`
}

func (m *PartnershipMetric) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// Known metric keys.
const (
	PartnershipMetricReach          = "reach"
	PartnershipMetricImpressions    = "impressions"
	PartnershipMetricClicks         = "clicks"
	PartnershipMetricPromoCodeUses  = "promo_code_uses"
	PartnershipMetricTicketsSold    = "tickets_sold"
	PartnershipMetricEstimatedSales = "estimated_sales"
	PartnershipMetricEngagement     = "engagement"
	PartnershipMetricCustom         = "custom"
)

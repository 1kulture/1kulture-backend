package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PartnershipType string

const (
	PartnershipTypeCash       PartnershipType = "cash_sponsorship"
	PartnershipTypeProduct    PartnershipType = "product_sponsorship"
	PartnershipTypeGiveaway   PartnershipType = "giveaway"
	PartnershipTypeAffiliate  PartnershipType = "affiliate"
	PartnershipTypeActivation PartnershipType = "brand_activation"
	PartnershipTypeMedia      PartnershipType = "media_partnership"
	PartnershipTypeVenue      PartnershipType = "venue_partnership"
	PartnershipTypeOther      PartnershipType = "other"
)

type BrandBenefit string

const (
	BrandBenefitLogoPlacement   BrandBenefit = "logo_placement"
	BrandBenefitSocialPromotion BrandBenefit = "social_media_promotion"
	BrandBenefitBoothAccess     BrandBenefit = "booth_access"
	BrandBenefitStageMention    BrandBenefit = "stage_mention"
	BrandBenefitEmailFeature    BrandBenefit = "email_feature"
	BrandBenefitProductSampling BrandBenefit = "product_sampling"
	BrandBenefitContentCollab   BrandBenefit = "content_collaboration"
	BrandBenefitOther           BrandBenefit = "other"
)

// EventPartnershipConfig is the "Open to Brand Partnerships" configuration for an event.
// One-to-one with Event.
type EventPartnershipConfig struct {
	BaseModel
	EventID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"event_id"`

	IsOpen bool `gorm:"default:false;index" json:"is_open"`

	// What the organizer is looking for
	PartnershipTypes datatypes.JSON `gorm:"type:jsonb" json:"partnership_types"` // []PartnershipType

	// What the brand gets
	BrandBenefits datatypes.JSON `gorm:"type:jsonb" json:"brand_benefits"` // []BrandBenefit

	// Audience breakdown shown to brands
	ExpectedAttendees int            `gorm:"default:0" json:"expected_attendees"`
	AudienceAgeRange  string         `gorm:"size:50" json:"audience_age_range"`    // "18-34"
	AudienceLocations datatypes.JSON `gorm:"type:jsonb" json:"audience_locations"` // []string
	AudienceInterests datatypes.JSON `gorm:"type:jsonb" json:"audience_interests"` // []string

	Notes string `gorm:"type:text" json:"notes"` // extra notes for brands

	Event Event `gorm:"foreignKey:EventID" json:"-"`
}

func (c *EventPartnershipConfig) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

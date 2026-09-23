package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BrandSize string

const (
	BrandSizeStartup    BrandSize = "startup"
	BrandSizeSmall      BrandSize = "small"
	BrandSizeMedium     BrandSize = "medium"
	BrandSizeLarge      BrandSize = "large"
	BrandSizeEnterprise BrandSize = "enterprise"
)

// BrandProfile is the business identity of a user with the brand role.
type BrandProfile struct {
	BaseModel
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`

	BusinessName string `gorm:"not null;size:255" json:"business_name"`
	LogoURL      string `gorm:"size:500" json:"logo_url"`
	Industry     string `gorm:"size:100;index" json:"industry"` // fashion | food | fintech | telecom | lifestyle | beauty | other
	Description  string `gorm:"type:text" json:"description"`
	Website      string `gorm:"size:500" json:"website"`
	Instagram    string `gorm:"size:100" json:"instagram"`
	Location     string `gorm:"size:255;index" json:"location"`
	ContactName  string `gorm:"size:255" json:"contact_name"`
	ContactEmail string `gorm:"size:255" json:"contact_email"`
	ContactPhone string `gorm:"size:30" json:"contact_phone"`

	BusinessSize BrandSize `gorm:"size:20;default:'small'" json:"business_size"`

	// Optional fields
	TargetAudience      string         `gorm:"type:text" json:"target_audience,omitempty"`
	PreferredCategories datatypes.JSON `gorm:"type:jsonb" json:"preferred_categories,omitempty"` // []string
	Tags                datatypes.JSON `gorm:"type:jsonb" json:"tags,omitempty"`                 // []string

	IsActive bool `gorm:"default:true;index" json:"is_active"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (b *BrandProfile) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	if b.BusinessSize == "" {
		b.BusinessSize = BrandSizeSmall
	}
	return nil
}

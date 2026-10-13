package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EventCategory is the canonical list of event types users can choose from.
// Managed by admins, seeded on first run.
type EventCategory struct {
	BaseModel
	Slug      string `gorm:"uniqueIndex;not null;size:50" json:"slug"`   // "concert", "tech-events"
	Name      string `gorm:"not null;size:100" json:"name"`               // "Concerts"
	Icon      string `gorm:"size:100" json:"icon"`                        // icon name/emoji for app
	SortOrder int    `gorm:"default:0" json:"sort_order"`
	IsActive  bool   `gorm:"default:true;index" json:"is_active"`
}

func (c *EventCategory) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// UserInterest links a user to the categories they care about.
type UserInterest struct {
	BaseModel
	UserID     uuid.UUID `gorm:"type:uuid;not null;index:idx_user_interest,unique" json:"user_id"`
	CategoryID uuid.UUID `gorm:"type:uuid;not null;index:idx_user_interest,unique" json:"category_id"`

	User     User          `gorm:"foreignKey:UserID" json:"-"`
	Category EventCategory `gorm:"foreignKey:CategoryID" json:"category"`
}

func (ui *UserInterest) BeforeCreate(tx *gorm.DB) error {
	if ui.ID == uuid.Nil {
		ui.ID = uuid.New()
	}
	return nil
}
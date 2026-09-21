package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Currency struct {
	BaseModel
	Code           string `gorm:"uniqueIndex;not null;size:3" json:"code"` // ISO 4217
	Name           string `gorm:"size:100;not null" json:"name"`           // Nigerian Naira
	Symbol         string `gorm:"size:10" json:"symbol"`                   // ₦
	IsEnabled      bool   `gorm:"default:false;index" json:"is_enabled"`   // super admin only
	IsDefault      bool   `gorm:"default:false" json:"is_default"`
	MinAmountMinor int64  `gorm:"default:0" json:"min_amount_minor"` // smallest accepted order in minor units
}

func (c *Currency) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type CurrencyRequestStatus string

const (
	CurrencyRequestPending  CurrencyRequestStatus = "pending"
	CurrencyRequestApproved CurrencyRequestStatus = "approved"
	CurrencyRequestRejected CurrencyRequestStatus = "rejected"
)

// CurrencyRequest lets Event Managers and Vendors ask admins to enable a currency.
type CurrencyRequest struct {
	BaseModel
	RequestedBy  uuid.UUID             `gorm:"type:uuid;not null;index" json:"requested_by"`
	CurrencyCode string                `gorm:"size:3;not null" json:"currency_code"`
	Reason       string                `gorm:"size:500" json:"reason"`
	Status       CurrencyRequestStatus `gorm:"size:20;not null;default:'pending';index" json:"status"`
	ReviewedBy   *uuid.UUID            `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	ReviewNote   string                `gorm:"size:500" json:"review_note"`

	Requester User `gorm:"foreignKey:RequestedBy" json:"-"`
}

func (c *CurrencyRequest) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.Status == "" {
		c.Status = CurrencyRequestPending
	}
	return nil
}

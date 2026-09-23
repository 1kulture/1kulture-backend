package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderItem struct {
	BaseModel
	OrderID      uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	TicketTypeID uuid.UUID `gorm:"type:uuid;not null;index" json:"ticket_type_id"`

	Quantity       int   `gorm:"not null" json:"quantity"`
	UnitPriceMinor int64 `gorm:"not null" json:"unit_price_minor"`
	TotalMinor     int64 `gorm:"not null" json:"total_minor"`

	// Snapshot of the ticket type name at purchase
	TicketTypeName string `gorm:"size:255" json:"ticket_type_name"`

	Order      Order      `gorm:"foreignKey:OrderID" json:"-"`
	TicketType TicketType `gorm:"foreignKey:TicketTypeID" json:"-"`
}

func (i *OrderItem) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

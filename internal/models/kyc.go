package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KYCStatus string

const (
	KYCStatusPending  KYCStatus = "pending"
	KYCStatusApproved KYCStatus = "approved"
	KYCStatusRejected KYCStatus = "rejected"
)

type KYCVerification struct {
	BaseModel
	UserID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	RoleID            uuid.UUID  `gorm:"type:uuid;not null" json:"role_id"`
	Status            KYCStatus  `gorm:"not null;default:'pending';size:20;index" json:"status"`
	DocumentType      string     `gorm:"size:50;not null" json:"document_type"`
	DocumentNumber    string     `gorm:"size:100;not null" json:"document_number"`
	DocumentURL       string     `gorm:"size:500;not null" json:"document_url"`
	SelfieURL         string     `gorm:"size:500;not null" json:"selfie_url"`
	AddressProofURL   string     `gorm:"size:500;not null" json:"address_proof_url"`
	BusinessName      string     `gorm:"size:255" json:"business_name"`
	BusinessRegNumber string     `gorm:"size:100" json:"business_reg_number"`
	SubmittedAt       time.Time  `gorm:"not null" json:"submitted_at"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy        *uuid.UUID `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	RejectionReason   string     `gorm:"size:500" json:"rejection_reason,omitempty"`
	User              User       `gorm:"foreignKey:UserID" json:"-"`
	Role              Role       `gorm:"foreignKey:RoleID" json:"-"`
}

func (k *KYCVerification) BeforeCreate(tx *gorm.DB) error {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	if k.SubmittedAt.IsZero() {
		k.SubmittedAt = time.Now()
	}
	return nil
}

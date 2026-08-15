package responses

import (
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID              uuid.UUID      `json:"id"`
	Email           string         `json:"email"`
	FirstName       string         `json:"first_name"`
	LastName        string         `json:"last_name"`
	PhoneNumber     string         `json:"phone_number,omitempty"`
	AvatarURL       string         `json:"avatar_url,omitempty"`
	Status          string         `json:"status"`
	EmailVerifiedAt *time.Time     `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time     `json:"last_login_at,omitempty"`
	Roles           []RoleResponse `json:"roles,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

type KYCResponse struct {
	ID                uuid.UUID  `json:"id"`
	UserID            uuid.UUID  `json:"user_id"`
	RoleID            uuid.UUID  `json:"role_id"`
	RoleName          string     `json:"role_name"`
	Status            string     `json:"status"`
	DocumentType      string     `json:"document_type"`
	DocumentNumber    string     `json:"document_number"`
	DocumentURL       string     `json:"document_url"`
	SelfieURL         string     `json:"selfie_url"`
	AddressProofURL   string     `json:"address_proof_url"`
	BusinessName      string     `json:"business_name,omitempty"`
	BusinessRegNumber string     `json:"business_reg_number,omitempty"`
	SubmittedAt       time.Time  `json:"submitted_at"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	RejectionReason   string     `json:"rejection_reason,omitempty"`
}

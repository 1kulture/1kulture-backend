package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserStatus string

const (
	UserStatusPending   UserStatus = "pending"
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

type User struct {
	BaseModel
	Email            string            `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash     string            `gorm:"not null;size:255" json:"-"`
	FirstName        string            `gorm:"size:100" json:"first_name"`
	LastName         string            `gorm:"size:100" json:"last_name"`
	PhoneNumber      string            `gorm:"size:20" json:"phone_number"`
	AvatarURL        string            `gorm:"size:500" json:"avatar_url"`
	Status           UserStatus        `gorm:"not null;default:'pending';size:20;index" json:"status"`
	EmailVerifiedAt  *time.Time        `json:"email_verified_at"`
	PhoneVerifiedAt  *time.Time        `json:"phone_verified_at"`
	LastLoginAt      *time.Time        `json:"last_login_at"`
	TwoFactorEnabled bool              `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret  string            `gorm:"size:100" json:"-"`
	Roles            []Role            `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	KYCVerifications []KYCVerification `gorm:"foreignKey:UserID" json:"kyc_verifications,omitempty"`
	RefreshTokens    []RefreshToken    `gorm:"foreignKey:UserID" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.Status == "" {
		u.Status = UserStatusPending
	}
	return nil
}

func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

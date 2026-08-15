package responses

import (
	"time"
)

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	TokenType    string       `json:"token_type"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type EmailVerificationResponse struct {
	Message    string    `json:"message"`
	Email      string    `json:"email"`
	VerifiedAt time.Time `json:"verified_at"`
}

type PasswordResetResponse struct {
	Message string `json:"message"`
}

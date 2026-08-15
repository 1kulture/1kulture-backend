package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
)

type AuthService interface {
	SignUp(ctx context.Context, req *requests.SignUpRequest) (*responses.AuthResponse, error)
	SignIn(ctx context.Context, req *requests.SignInRequest) (*responses.AuthResponse, error)
	VerifyEmail(ctx context.Context, req *requests.VerifyEmailRequest) error
	ResendVerification(ctx context.Context, req *requests.ResendVerificationRequest) error
	RefreshToken(ctx context.Context, req *requests.RefreshTokenRequest) (*responses.TokenResponse, error)
	Logout(ctx context.Context, req *requests.LogoutRequest) error
	ForgotPassword(ctx context.Context, req *requests.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *requests.ResetPasswordRequest) error
	ChangePassword(ctx context.Context, userID string, req *requests.ChangePasswordRequest) error
}

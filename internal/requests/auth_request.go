package requests

type SignUpRequest struct {
	Email       string `json:"email" validate:"required,email,max=255"`
	Password    string `json:"password" validate:"required,min=8,max=72"`
	FirstName   string `json:"first_name" validate:"required,min=2,max=100"`
	LastName    string `json:"last_name" validate:"required,min=2,max=100"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,min=10,max=20"`
	Role        string `json:"role" validate:"omitempty,oneof=guest vendor event_manager"`
	AcceptTerms bool   `json:"accept_terms" validate:"required,eq=true"`
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
	Code  string `json:"code" validate:"required,len=6,numeric"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,jwt"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=8,max=72"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=72,nefield=CurrentPassword"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,jwt"`
}

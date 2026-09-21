package requests

// CreateCurrencyRequest - Super Admin only.
type CreateCurrencyRequest struct {
	Code           string `json:"code" validate:"required,len=3,uppercase" example:"NGN"`
	Name           string `json:"name" validate:"required,min=2,max=100" example:"Nigerian Naira"`
	Symbol         string `json:"symbol" validate:"omitempty,max=10" example:"₦"`
	IsEnabled      bool   `json:"is_enabled"`
	IsDefault      bool   `json:"is_default"`
	MinAmountMinor int64  `json:"min_amount_minor" validate:"omitempty,min=0"`
}

// UpdateCurrencyRequest - Super Admin only.
type UpdateCurrencyRequest struct {
	Name           string `json:"name" validate:"omitempty,min=2,max=100"`
	Symbol         string `json:"symbol" validate:"omitempty,max=10"`
	IsEnabled      *bool  `json:"is_enabled,omitempty"`
	IsDefault      *bool  `json:"is_default,omitempty"`
	MinAmountMinor *int64 `json:"min_amount_minor,omitempty" validate:"omitempty,min=0"`
}

// RequestCurrencyRequest - Event Manager or Vendor.
type RequestCurrencyRequest struct {
	CurrencyCode string `json:"currency_code" validate:"required,len=3,uppercase" example:"USD"`
	Reason       string `json:"reason" validate:"required,min=10,max=500" example:"We want to accept international ticket sales."`
}

// ReviewCurrencyRequestRequest - Super Admin decision.
type ReviewCurrencyRequestRequest struct {
	Status     string `json:"status" validate:"required,oneof=approved rejected" example:"approved"`
	ReviewNote string `json:"review_note" validate:"omitempty,max=500"`
}

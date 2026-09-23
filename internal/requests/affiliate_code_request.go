package requests

// CreateAffiliateCodeRequest creates a promo code tied to a partnership.
type CreateAffiliateCodeRequest struct {
	Code string `json:"code" validate:"required,min=3,max=50,uppercase,alphanumeric"`

	// Promo semantics
	Type       string `json:"type" validate:"required,oneof=percentage fixed"`
	ValueMinor int64  `json:"value_minor" validate:"required,min=1"`

	// Commission the brand earns per conversion (in bps of the order total)
	BrandCommissionBps int `json:"brand_commission_bps" validate:"required,min=0,max=5000"`

	UsageLimit   int `json:"usage_limit" validate:"omitempty,min=0"`
	PerUserLimit int `json:"per_user_limit" validate:"omitempty,min=0"`
}

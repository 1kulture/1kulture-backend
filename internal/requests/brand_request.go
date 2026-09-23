package requests

// CreateBrandProfileRequest creates or updates the brand profile.
type CreateBrandProfileRequest struct {
	BusinessName string `json:"business_name" validate:"required,min=2,max=255"`
	LogoURL      string `json:"logo_url" validate:"omitempty,url,max=500"`
	Industry     string `json:"industry" validate:"required,min=2,max=100"`
	Description  string `json:"description" validate:"omitempty,max=2000"`
	Website      string `json:"website" validate:"omitempty,url,max=500"`
	Instagram    string `json:"instagram" validate:"omitempty,max=100"`
	Location     string `json:"location" validate:"required,min=2,max=255"`
	ContactName  string `json:"contact_name" validate:"required,min=2,max=255"`
	ContactEmail string `json:"contact_email" validate:"omitempty,email,max=255"`
	ContactPhone string `json:"contact_phone" validate:"omitempty,min=7,max=30"`

	BusinessSize string `json:"business_size" validate:"omitempty,oneof=startup small medium large enterprise"`

	TargetAudience      string   `json:"target_audience" validate:"omitempty,max=2000"`
	PreferredCategories []string `json:"preferred_categories" validate:"omitempty,max=20,dive,min=1,max=100"`
	Tags                []string `json:"tags" validate:"omitempty,max=20,dive,min=1,max=50"`
}

// UpdateBrandProfileRequest allows partial updates.
type UpdateBrandProfileRequest struct {
	BusinessName *string `json:"business_name,omitempty" validate:"omitempty,min=2,max=255"`
	LogoURL      *string `json:"logo_url,omitempty" validate:"omitempty,url,max=500"`
	Industry     *string `json:"industry,omitempty" validate:"omitempty,min=2,max=100"`
	Description  *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	Website      *string `json:"website,omitempty" validate:"omitempty,url,max=500"`
	Instagram    *string `json:"instagram,omitempty" validate:"omitempty,max=100"`
	Location     *string `json:"location,omitempty" validate:"omitempty,min=2,max=255"`
	ContactName  *string `json:"contact_name,omitempty" validate:"omitempty,min=2,max=255"`
	ContactEmail *string `json:"contact_email,omitempty" validate:"omitempty,email,max=255"`
	ContactPhone *string `json:"contact_phone,omitempty" validate:"omitempty,min=7,max=30"`
	BusinessSize *string `json:"business_size,omitempty" validate:"omitempty,oneof=startup small medium large enterprise"`

	TargetAudience      *string  `json:"target_audience,omitempty" validate:"omitempty,max=2000"`
	PreferredCategories []string `json:"preferred_categories,omitempty" validate:"omitempty,max=20,dive,min=1,max=100"`
	Tags                []string `json:"tags,omitempty" validate:"omitempty,max=20,dive,min=1,max=50"`

	IsActive *bool `json:"is_active,omitempty"`
}

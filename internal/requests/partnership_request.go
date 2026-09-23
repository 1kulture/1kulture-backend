package requests

// EventPartnershipConfigRequest configures an event's brand partnership mode.
type EventPartnershipConfigRequest struct {
	IsOpen bool `json:"is_open"`

	PartnershipTypes []string `json:"partnership_types" validate:"omitempty,max=10,dive,oneof=cash_sponsorship product_sponsorship giveaway affiliate brand_activation media_partnership venue_partnership other"`
	BrandBenefits    []string `json:"brand_benefits" validate:"omitempty,max=10,dive,oneof=logo_placement social_media_promotion booth_access stage_mention email_feature product_sampling content_collaboration other"`

	ExpectedAttendees int      `json:"expected_attendees" validate:"omitempty,min=0"`
	AudienceAgeRange  string   `json:"audience_age_range" validate:"omitempty,max=50"`
	AudienceLocations []string `json:"audience_locations" validate:"omitempty,max=20,dive,min=1,max=100"`
	AudienceInterests []string `json:"audience_interests" validate:"omitempty,max=20,dive,min=1,max=100"`

	Notes string `json:"notes" validate:"omitempty,max=3000"`
}

// OpportunityCreateRequest creates an opportunity inside an event config.
type OpportunityCreateRequest struct {
	Title       string `json:"title" validate:"required,min=2,max=255"`
	Description string `json:"description" validate:"omitempty,max=3000"`
	Type        string `json:"type" validate:"required,oneof=cash_sponsorship product_sponsorship giveaway affiliate brand_activation media_partnership venue_partnership other"`

	BudgetMinMinor int64 `json:"budget_min_minor" validate:"omitempty,min=0"`
	BudgetMaxMinor int64 `json:"budget_max_minor" validate:"omitempty,min=0"`

	SlotsTotal int   `json:"slots_total" validate:"required,min=1"`
	SortOrder  int   `json:"sort_order" validate:"omitempty,min=0"`
	IsActive   *bool `json:"is_active,omitempty"`
}

// OpportunityUpdateRequest updates an opportunity.
type OpportunityUpdateRequest struct {
	Title          *string `json:"title,omitempty" validate:"omitempty,min=2,max=255"`
	Description    *string `json:"description,omitempty" validate:"omitempty,max=3000"`
	Type           *string `json:"type,omitempty" validate:"omitempty,oneof=cash_sponsorship product_sponsorship giveaway affiliate brand_activation media_partnership venue_partnership other"`
	BudgetMinMinor *int64  `json:"budget_min_minor,omitempty" validate:"omitempty,min=0"`
	BudgetMaxMinor *int64  `json:"budget_max_minor,omitempty" validate:"omitempty,min=0"`
	SlotsTotal     *int    `json:"slots_total,omitempty" validate:"omitempty,min=1"`
	SortOrder      *int    `json:"sort_order,omitempty" validate:"omitempty,min=0"`
	IsActive       *bool   `json:"is_active,omitempty"`
}

// PartnershipRequestCreate is a brand's request to partner with an event.
type PartnershipRequestCreate struct {
	OpportunityID  string   `json:"opportunity_id,omitempty" validate:"omitempty,uuid"`
	RequestedTypes []string `json:"requested_types" validate:"required,min=1,max=10,dive,oneof=cash_sponsorship product_sponsorship giveaway affiliate brand_activation media_partnership venue_partnership other"`
	Message        string   `json:"message" validate:"required,min=10,max=2000"`
}

// PartnershipDeclineRequest is used by organizers to decline or cancel.
type PartnershipDeclineRequest struct {
	Reason string `json:"reason" validate:"omitempty,max=500"`
}

// PartnershipMetricRequest is used by both sides to report metrics.
type PartnershipMetricRequest struct {
	Key      string `json:"key" validate:"required,min=2,max=50"`
	Value    int64  `json:"value" validate:"required"`
	Currency string `json:"currency" validate:"omitempty,len=3,uppercase"`
	Note     string `json:"note" validate:"omitempty,max=500"`
}

package requests

// UpdateSettingRequest updates a single setting by key.
type UpdateSettingRequest struct {
	Key   string `json:"key" validate:"required,min=2,max=100" example:"escrow.release_days"`
	Value string `json:"value" validate:"required" example:"3"`
}

// BulkUpdateSettingsRequest updates multiple settings at once.
type BulkUpdateSettingsRequest struct {
	Settings []UpdateSettingRequest `json:"settings" validate:"required,min=1,dive"`
}

// CreateSettingRequest is for admins adding a brand-new setting.
type CreateSettingRequest struct {
	Key         string `json:"key" validate:"required,min=2,max=100" example:"custom.setting"`
	Value       string `json:"value" validate:"required"`
	ValueType   string `json:"value_type" validate:"required,oneof=string int bool json duration" example:"string"`
	Category    string `json:"category" validate:"omitempty,max=50" example:"general"`
	Description string `json:"description" validate:"omitempty,max=500"`
	IsPublic    bool   `json:"is_public"`
}

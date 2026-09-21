package responses

import (
	"time"

	"github.com/google/uuid"
)

type SettingResponse struct {
	ID          uuid.UUID  `json:"id"`
	Key         string     `json:"key"`
	Value       string     `json:"value"`
	ValueType   string     `json:"value_type"`
	Category    string     `json:"category"`
	Description string     `json:"description"`
	IsPublic    bool       `json:"is_public"`
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type PublicSettingsResponse struct {
	Settings map[string]string `json:"settings"`
}

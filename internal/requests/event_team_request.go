package requests

// CoOrganizerAddRequest adds a co-organizer to an event.
type CoOrganizerAddRequest struct {
	UserID      string                 `json:"user_id" validate:"required,uuid"`
	Role        string                 `json:"role" validate:"required,oneof=co_host manager staff"`
	Permissions map[string]interface{} `json:"permissions,omitempty"`
}

// StaffAddRequest assigns a staff member to an event.
type StaffAddRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	Role   string `json:"role" validate:"required,oneof=scanner checkin_lead coordinator"`
}

// ShareRequest records a share event.
type ShareRequest struct {
	Channel string `json:"channel" validate:"required,oneof=link whatsapp x facebook email other"`
}

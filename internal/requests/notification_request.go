package requests

// MarkNotificationReadRequest can target a single or all notifications.
type MarkNotificationReadRequest struct {
	NotificationID string `json:"notification_id,omitempty" validate:"omitempty,uuid"`
	MarkAll        bool   `json:"mark_all,omitempty"`
}

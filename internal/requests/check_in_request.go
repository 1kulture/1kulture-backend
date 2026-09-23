package requests

type CheckInRequest struct {
	QRPayload  string `json:"qr_payload" validate:"required,min=10"`
	DeviceInfo string `json:"device_info,omitempty" validate:"omitempty,max=255"`
	Location   string `json:"location,omitempty" validate:"omitempty,max=255"`
}

type ManualCheckInRequest struct {
	TicketCode string `json:"ticket_code" validate:"required,min=6,max=32"`
	DeviceInfo string `json:"device_info,omitempty" validate:"omitempty,max=255"`
	Location   string `json:"location,omitempty" validate:"omitempty,max=255"`
}

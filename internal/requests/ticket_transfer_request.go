package requests

type InitiateTransferRequest struct {
	ToEmail string `json:"to_email" validate:"required,email,max=255"`
	Message string `json:"message" validate:"omitempty,max=500"`
}

type AcceptTransferRequest struct {
	Token string `json:"token" validate:"required,min=20"`
}

type DeclineTransferRequest struct {
	Token string `json:"token" validate:"required,min=20"`
}

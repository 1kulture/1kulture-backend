package requests

type OrderItemInput struct {
	TicketTypeID string `json:"ticket_type_id" validate:"required,uuid"`
	Quantity     int    `json:"quantity" validate:"required,min=1,max=50"`
}

type CreateOrderRequest struct {
	EventID        string           `json:"event_id" validate:"required,uuid"`
	Items          []OrderItemInput `json:"items" validate:"required,min=1,max=10,dive"`
	PromoCode      string           `json:"promo_code,omitempty" validate:"omitempty,min=3,max=50"`
	IdempotencyKey string           `json:"idempotency_key" validate:"omitempty,max=100"`

	// Buyer details — optional; defaults to user profile
	BuyerName  string `json:"buyer_name" validate:"omitempty,min=2,max=255"`
	BuyerEmail string `json:"buyer_email" validate:"omitempty,email,max=255"`
	BuyerPhone string `json:"buyer_phone" validate:"omitempty,min=10,max=20"`
}

type InitializePaymentRequest struct {
	// Optional: callback URL for redirect after payment
	CallbackURL string `json:"callback_url,omitempty" validate:"omitempty,url,max=500"`
}

type RefundRequestPayload struct {
	TicketIDs []string `json:"ticket_ids" validate:"required,min=1,dive,uuid"`
	Reason    string   `json:"reason" validate:"required,min=5,max=500"`
}

type RefundDecisionRequest struct {
	Status     string `json:"status" validate:"required,oneof=approved rejected"`
	ReviewNote string `json:"review_note" validate:"omitempty,max=500"`
}

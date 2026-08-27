package requests

type WaitlistRequest struct {
	Email    string `json:"email" validate:"required,email,max=255" example:"user@example.com"`
	Category string `json:"category" validate:"required,min=2,max=100" example:"event_planner"`
}

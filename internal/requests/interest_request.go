package requests

type UpdateInterestsRequest struct {
	// Full replacement — send all category slugs the user wants.
	// Send empty array to clear all interests.
	CategorySlugs []string `json:"category_slugs" validate:"max=30,dive,min=1,max=50"`
}
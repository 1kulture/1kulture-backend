package responses

import (
	"github.com/google/uuid"
)

type EventCategoryResponse struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon,omitempty"`
	SortOrder int       `json:"sort_order"`
}

type UserInterestsResponse struct {
	Interests []EventCategoryResponse `json:"interests"`
	Count     int                     `json:"count"`
}

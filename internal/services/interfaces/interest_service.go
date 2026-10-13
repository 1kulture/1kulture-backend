package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
)

type InterestService interface {
	// ListCategories returns all active categories (used to render the picker).
	ListCategories(ctx context.Context) ([]responses.EventCategoryResponse, error)

	// GetMyInterests returns the current user's selected categories.
	GetMyInterests(ctx context.Context, userID uuid.UUID) (*responses.UserInterestsResponse, error)

	// UpdateMyInterests replaces the user's interests with the given set of slugs.
	// Send empty list to clear all interests.
	UpdateMyInterests(ctx context.Context, userID uuid.UUID, req *requests.UpdateInterestsRequest) (*responses.UserInterestsResponse, error)

	// GetInterestSlugs returns just the slugs (used by event filtering logic).
	GetInterestSlugs(ctx context.Context, userID uuid.UUID) ([]string, error)
}
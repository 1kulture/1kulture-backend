package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/1kulture/1kulture-backend/internal/models"
)

type EventCategoryRepository interface {
	Create(ctx context.Context, c *models.EventCategory) error
	FindBySlug(ctx context.Context, slug string) (*models.EventCategory, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.EventCategory, error)
	FindBySlugs(ctx context.Context, slugs []string) ([]models.EventCategory, error)
	ListActive(ctx context.Context) ([]models.EventCategory, error)
	Update(ctx context.Context, c *models.EventCategory) error
}

type UserInterestRepository interface {
	ReplaceForUser(ctx context.Context, userID uuid.UUID, categoryIDs []uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]models.UserInterest, error)
	ListCategoryIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}
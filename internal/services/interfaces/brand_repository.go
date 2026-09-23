package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type BrandProfileRepository interface {
	Create(ctx context.Context, profile *models.BrandProfile) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.BrandProfile, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.BrandProfile, error)
	Update(ctx context.Context, profile *models.BrandProfile) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, industry, location string, page, perPage int) ([]models.BrandProfile, int64, error)
}

package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type CommissionTierRepository interface {
	Create(ctx context.Context, tier *models.CommissionTier) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.CommissionTier, error)
	FindAll(ctx context.Context, onlyActive bool) ([]models.CommissionTier, error)
	Update(ctx context.Context, tier *models.CommissionTier) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type OrganizerCommissionOverrideRepository interface {
	Upsert(ctx context.Context, override *models.OrganizerCommissionOverride) error
	FindByOrganizer(ctx context.Context, organizerID uuid.UUID) (*models.OrganizerCommissionOverride, error)
	Delete(ctx context.Context, organizerID uuid.UUID) error
}

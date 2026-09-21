package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type EventOccurrenceRepository interface {
	Create(ctx context.Context, occ *models.EventOccurrence) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.EventOccurrence, error)
	FindByEvent(ctx context.Context, eventID uuid.UUID) ([]models.EventOccurrence, error)
	Update(ctx context.Context, occ *models.EventOccurrence) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByEvent(ctx context.Context, eventID uuid.UUID) error
}

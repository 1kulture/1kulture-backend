package interfaces

import (
	"context"
	"time"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type EventListFilter struct {
	OrganizerID *uuid.UUID
	Category    string
	City        string
	Country     string
	Status      string
	EventType   string
	Search      string
	StartFrom   *time.Time
	StartTo     *time.Time
	Featured    *bool
	Page        int
	PerPage     int
}

type EventRepository interface {
	Create(ctx context.Context, event *models.Event) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Event, error)
	FindBySlug(ctx context.Context, slug string) (*models.Event, error)
	FindWithRelations(ctx context.Context, id uuid.UUID) (*models.Event, error)
	Update(ctx context.Context, event *models.Event) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter EventListFilter) ([]models.Event, int64, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	IncrementField(ctx context.Context, id uuid.UUID, field string, delta int) error
}

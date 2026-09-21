package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type eventOccurrenceRepository struct {
	db *gorm.DB
}

func NewEventOccurrenceRepository(db *gorm.DB) interfaces.EventOccurrenceRepository {
	return &eventOccurrenceRepository{db: db}
}

func (r *eventOccurrenceRepository) Create(ctx context.Context, occ *models.EventOccurrence) error {
	if err := r.db.WithContext(ctx).Create(occ).Error; err != nil {
		return fmt.Errorf("failed to create occurrence: %w", err)
	}
	return nil
}

func (r *eventOccurrenceRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.EventOccurrence, error) {
	var o models.EventOccurrence
	if err := r.db.WithContext(ctx).First(&o, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find occurrence: %w", err)
	}
	return &o, nil
}

func (r *eventOccurrenceRepository) FindByEvent(ctx context.Context, eventID uuid.UUID) ([]models.EventOccurrence, error) {
	var list []models.EventOccurrence
	if err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Order("sort_order asc, start_at asc").
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to list occurrences: %w", err)
	}
	return list, nil
}

func (r *eventOccurrenceRepository) Update(ctx context.Context, occ *models.EventOccurrence) error {
	if err := r.db.WithContext(ctx).Save(occ).Error; err != nil {
		return fmt.Errorf("failed to update occurrence: %w", err)
	}
	return nil
}

func (r *eventOccurrenceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.EventOccurrence{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete occurrence: %w", err)
	}
	return nil
}

func (r *eventOccurrenceRepository) DeleteByEvent(ctx context.Context, eventID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Delete(&models.EventOccurrence{}).Error; err != nil {
		return fmt.Errorf("failed to delete occurrences: %w", err)
	}
	return nil
}

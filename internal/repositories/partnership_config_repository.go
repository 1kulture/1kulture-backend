package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type eventPartnershipConfigRepository struct {
	db *gorm.DB
}

func NewEventPartnershipConfigRepository(db *gorm.DB) repoInterfaces.EventPartnershipConfigRepository {
	return &eventPartnershipConfigRepository{db: db}
}

func (r *eventPartnershipConfigRepository) Create(ctx context.Context, cfg *models.EventPartnershipConfig) error {
	if err := r.db.WithContext(ctx).Create(cfg).Error; err != nil {
		return fmt.Errorf("failed to create partnership config: %w", err)
	}
	return nil
}

func (r *eventPartnershipConfigRepository) FindByEventID(ctx context.Context, eventID uuid.UUID) (*models.EventPartnershipConfig, error) {
	var cfg models.EventPartnershipConfig
	if err := r.db.WithContext(ctx).First(&cfg, "event_id = ?", eventID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find partnership config: %w", err)
	}
	return &cfg, nil
}

func (r *eventPartnershipConfigRepository) Update(ctx context.Context, cfg *models.EventPartnershipConfig) error {
	if err := r.db.WithContext(ctx).Save(cfg).Error; err != nil {
		return fmt.Errorf("failed to update partnership config: %w", err)
	}
	return nil
}

func (r *eventPartnershipConfigRepository) Delete(ctx context.Context, eventID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Delete(&models.EventPartnershipConfig{}).Error; err != nil {
		return fmt.Errorf("failed to delete partnership config: %w", err)
	}
	return nil
}

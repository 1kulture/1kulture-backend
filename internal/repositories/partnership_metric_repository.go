package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type partnershipMetricRepository struct {
	db *gorm.DB
}

func NewPartnershipMetricRepository(db *gorm.DB) repoInterfaces.PartnershipMetricRepository {
	return &partnershipMetricRepository{db: db}
}

func (r *partnershipMetricRepository) Create(ctx context.Context, m *models.PartnershipMetric) error {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("failed to create partnership metric: %w", err)
	}
	return nil
}

func (r *partnershipMetricRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.PartnershipMetric, error) {
	var m models.PartnershipMetric
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find partnership metric: %w", err)
	}
	return &m, nil
}

func (r *partnershipMetricRepository) FindByPartnership(ctx context.Context, partnershipID uuid.UUID) ([]models.PartnershipMetric, error) {
	var list []models.PartnershipMetric
	if err := r.db.WithContext(ctx).
		Where("partnership_id = ?", partnershipID).
		Order("created_at desc").
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to list partnership metrics: %w", err)
	}
	return list, nil
}

func (r *partnershipMetricRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.PartnershipMetric{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete partnership metric: %w", err)
	}
	return nil
}

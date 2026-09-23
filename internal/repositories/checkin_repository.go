package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type checkInRepository struct {
	db *gorm.DB
}

func NewCheckInRepository(db *gorm.DB) repoInterfaces.CheckInRepository {
	return &checkInRepository{db: db}
}

func (r *checkInRepository) Create(ctx context.Context, ci *models.CheckIn) error {
	if err := r.db.WithContext(ctx).Create(ci).Error; err != nil {
		return fmt.Errorf("failed to create check-in: %w", err)
	}
	return nil
}

func (r *checkInRepository) FindByTicket(ctx context.Context, ticketID uuid.UUID) (*models.CheckIn, error) {
	var ci models.CheckIn
	if err := r.db.WithContext(ctx).
		Where("ticket_id = ?", ticketID).
		Order("scanned_at desc").
		First(&ci).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find check-in: %w", err)
	}
	return &ci, nil
}

func (r *checkInRepository) FindByEvent(ctx context.Context, eventID uuid.UUID, page, perPage int) ([]models.CheckIn, int64, error) {
	var list []models.CheckIn
	var total int64
	q := r.db.WithContext(ctx).Model(&models.CheckIn{}).Where("event_id = ?", eventID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count check-ins: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Order("scanned_at desc").Limit(perPage).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list check-ins: %w", err)
	}
	return list, total, nil
}

func (r *checkInRepository) CountByEvent(ctx context.Context, eventID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.CheckIn{}).
		Where("event_id = ?", eventID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count check-ins: %w", err)
	}
	return count, nil
}

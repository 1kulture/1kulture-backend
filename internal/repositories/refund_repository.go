package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type refundRepository struct {
	db *gorm.DB
}

func NewRefundRepository(db *gorm.DB) repoInterfaces.RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(ctx context.Context, refund *models.Refund) error {
	if err := r.db.WithContext(ctx).Create(refund).Error; err != nil {
		return fmt.Errorf("failed to create refund: %w", err)
	}
	return nil
}

func (r *refundRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Refund, error) {
	var refund models.Refund
	if err := r.db.WithContext(ctx).First(&refund, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find refund: %w", err)
	}
	return &refund, nil
}

func (r *refundRepository) FindByOrder(ctx context.Context, orderID uuid.UUID) ([]models.Refund, error) {
	var list []models.Refund
	if err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at desc").
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to list refunds by order: %w", err)
	}
	return list, nil
}

func (r *refundRepository) Update(ctx context.Context, refund *models.Refund) error {
	if err := r.db.WithContext(ctx).Save(refund).Error; err != nil {
		return fmt.Errorf("failed to update refund: %w", err)
	}
	return nil
}

func (r *refundRepository) ListByStatus(ctx context.Context, status string, page, perPage int) ([]models.Refund, int64, error) {
	var list []models.Refund
	var total int64
	q := r.db.WithContext(ctx).Model(&models.Refund{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count refunds: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Order("created_at desc").Limit(perPage).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list refunds: %w", err)
	}
	return list, total, nil
}

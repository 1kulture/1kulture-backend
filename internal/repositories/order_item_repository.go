package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type orderItemRepository struct {
	db *gorm.DB
}

func NewOrderItemRepository(db *gorm.DB) repoInterfaces.OrderItemRepository {
	return &orderItemRepository{db: db}
}

func (r *orderItemRepository) CreateMany(ctx context.Context, items []models.OrderItem) error {
	if len(items) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&items).Error; err != nil {
		return fmt.Errorf("failed to create order items: %w", err)
	}
	return nil
}

func (r *orderItemRepository) FindByOrder(ctx context.Context, orderID uuid.UUID) ([]models.OrderItem, error) {
	var list []models.OrderItem
	if err := r.db.WithContext(ctx).
		Preload("TicketType").
		Where("order_id = ?", orderID).
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to find order items: %w", err)
	}
	return list, nil
}

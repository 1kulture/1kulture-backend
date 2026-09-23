package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) repoInterfaces.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *models.Order) error {
	if err := r.db.WithContext(ctx).Create(order).Error; err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

func (r *orderRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	var o models.Order
	if err := r.db.WithContext(ctx).First(&o, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find order: %w", err)
	}
	return &o, nil
}

func (r *orderRepository) FindByReference(ctx context.Context, reference string) (*models.Order, error) {
	var o models.Order
	if err := r.db.WithContext(ctx).First(&o, "reference = ?", reference).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find order by reference: %w", err)
	}
	return &o, nil
}

func (r *orderRepository) FindByIdempotencyKey(ctx context.Context, key string) (*models.Order, error) {
	if key == "" {
		return nil, nil
	}
	var o models.Order
	if err := r.db.WithContext(ctx).First(&o, "idempotency_key = ?", key).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find order by idempotency key: %w", err)
	}
	return &o, nil
}

func (r *orderRepository) FindWithItems(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	var o models.Order
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Tickets").
		First(&o, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find order with items: %w", err)
	}
	return &o, nil
}

func (r *orderRepository) FindWithTickets(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	var o models.Order
	if err := r.db.WithContext(ctx).
		Preload("Tickets").
		Preload("Tickets.TicketType").
		Preload("Items").
		First(&o, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find order with tickets: %w", err)
	}
	return &o, nil
}

func (r *orderRepository) Update(ctx context.Context, order *models.Order) error {
	if err := r.db.WithContext(ctx).Save(order).Error; err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}
	return nil
}

func (r *orderRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Order, int64, error) {
	var list []models.Order
	var total int64
	q := r.db.WithContext(ctx).Model(&models.Order{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Preload("Items").Preload("Tickets").
		Order("created_at desc").
		Limit(perPage).Offset(offset).
		Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}
	return list, total, nil
}

func (r *orderRepository) ListByEvent(ctx context.Context, eventID uuid.UUID, page, perPage int) ([]models.Order, int64, error) {
	var list []models.Order
	var total int64
	q := r.db.WithContext(ctx).Model(&models.Order{}).Where("event_id = ?", eventID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Preload("Items").
		Order("created_at desc").
		Limit(perPage).Offset(offset).
		Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list event orders: %w", err)
	}
	return list, total, nil
}

func (r *orderRepository) FindExpiredPending(ctx context.Context, before time.Time, limit int) ([]models.Order, error) {
	var list []models.Order
	if err := r.db.WithContext(ctx).
		Where("status = ? AND expires_at IS NOT NULL AND expires_at < ?", models.OrderStatusPending, before).
		Order("expires_at asc").
		Limit(limit).
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to find expired orders: %w", err)
	}
	return list, nil
}

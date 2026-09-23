package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type ticketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) repoInterfaces.TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) CreateMany(ctx context.Context, tickets []models.Ticket) error {
	if len(tickets) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&tickets).Error; err != nil {
		return fmt.Errorf("failed to create tickets: %w", err)
	}
	return nil
}

func (r *ticketRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Ticket, error) {
	var t models.Ticket
	if err := r.db.WithContext(ctx).
		Preload("TicketType").
		Preload("Event").
		First(&t, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find ticket: %w", err)
	}
	return &t, nil
}

func (r *ticketRepository) FindByCode(ctx context.Context, code string) (*models.Ticket, error) {
	var t models.Ticket
	if err := r.db.WithContext(ctx).
		Preload("TicketType").
		Preload("Event").
		First(&t, "code = ?", code).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find ticket by code: %w", err)
	}
	return &t, nil
}

func (r *ticketRepository) FindByOrder(ctx context.Context, orderID uuid.UUID) ([]models.Ticket, error) {
	var list []models.Ticket
	if err := r.db.WithContext(ctx).
		Preload("TicketType").
		Where("order_id = ?", orderID).
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to find tickets by order: %w", err)
	}
	return list, nil
}

func (r *ticketRepository) FindByUser(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Ticket, int64, error) {
	var list []models.Ticket
	var total int64
	q := r.db.WithContext(ctx).Model(&models.Ticket{}).Where("user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tickets: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Preload("TicketType").Preload("Event").
		Order("created_at desc").
		Limit(perPage).Offset(offset).
		Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list tickets: %w", err)
	}
	return list, total, nil
}

func (r *ticketRepository) FindByEvent(ctx context.Context, eventID uuid.UUID, page, perPage int) ([]models.Ticket, int64, error) {
	var list []models.Ticket
	var total int64
	q := r.db.WithContext(ctx).Model(&models.Ticket{}).Where("event_id = ?", eventID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count event tickets: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Preload("TicketType").
		Order("created_at desc").
		Limit(perPage).Offset(offset).
		Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list event tickets: %w", err)
	}
	return list, total, nil
}

func (r *ticketRepository) Update(ctx context.Context, ticket *models.Ticket) error {
	if err := r.db.WithContext(ctx).Save(ticket).Error; err != nil {
		return fmt.Errorf("failed to update ticket: %w", err)
	}
	return nil
}

func (r *ticketRepository) UpdateMany(ctx context.Context, tickets []models.Ticket) error {
	if len(tickets) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range tickets {
			if err := tx.Save(&tickets[i]).Error; err != nil {
				return fmt.Errorf("failed to update ticket %s: %w", tickets[i].ID, err)
			}
		}
		return nil
	})
}

func (r *ticketRepository) CountByEvent(ctx context.Context, eventID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Ticket{}).
		Where("event_id = ?", eventID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count tickets: %w", err)
	}
	return count, nil
}

func (r *ticketRepository) CountUsedByEvent(ctx context.Context, eventID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Ticket{}).
		Where("event_id = ? AND status = ?", eventID, models.TicketStatusUsed).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count used tickets: %w", err)
	}
	return count, nil
}

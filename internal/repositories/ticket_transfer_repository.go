package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type ticketTransferRepository struct {
	db *gorm.DB
}

func NewTicketTransferRepository(db *gorm.DB) repoInterfaces.TicketTransferRepository {
	return &ticketTransferRepository{db: db}
}

func (r *ticketTransferRepository) Create(ctx context.Context, t *models.TicketTransfer) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return fmt.Errorf("failed to create ticket transfer: %w", err)
	}
	return nil
}

func (r *ticketTransferRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.TicketTransfer, error) {
	var tr models.TicketTransfer
	if err := r.db.WithContext(ctx).Preload("Ticket").First(&tr, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find ticket transfer: %w", err)
	}
	return &tr, nil
}

func (r *ticketTransferRepository) FindByToken(ctx context.Context, token string) (*models.TicketTransfer, error) {
	var tr models.TicketTransfer
	if err := r.db.WithContext(ctx).Preload("Ticket").First(&tr, "token = ?", token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find ticket transfer by token: %w", err)
	}
	return &tr, nil
}

func (r *ticketTransferRepository) FindByTicket(ctx context.Context, ticketID uuid.UUID) ([]models.TicketTransfer, error) {
	var list []models.TicketTransfer
	if err := r.db.WithContext(ctx).
		Where("ticket_id = ?", ticketID).
		Order("created_at desc").
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to list ticket transfers: %w", err)
	}
	return list, nil
}

func (r *ticketTransferRepository) FindByUser(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.TicketTransfer, int64, error) {
	var list []models.TicketTransfer
	var total int64
	q := r.db.WithContext(ctx).Model(&models.TicketTransfer{}).Where("from_user_id = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count ticket transfers: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Preload("Ticket").Order("created_at desc").Limit(perPage).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list ticket transfers: %w", err)
	}
	return list, total, nil
}

func (r *ticketTransferRepository) Update(ctx context.Context, t *models.TicketTransfer) error {
	if err := r.db.WithContext(ctx).Save(t).Error; err != nil {
		return fmt.Errorf("failed to update ticket transfer: %w", err)
	}
	return nil
}

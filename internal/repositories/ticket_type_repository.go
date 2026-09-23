package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type ticketTypeRepository struct {
	db *gorm.DB
}

func NewTicketTypeRepository(db *gorm.DB) repoInterfaces.TicketTypeRepository {
	return &ticketTypeRepository{db: db}
}

func (r *ticketTypeRepository) Create(ctx context.Context, tt *models.TicketType) error {
	if err := r.db.WithContext(ctx).Create(tt).Error; err != nil {
		return fmt.Errorf("failed to create ticket type: %w", err)
	}
	return nil
}

func (r *ticketTypeRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.TicketType, error) {
	var tt models.TicketType
	if err := r.db.WithContext(ctx).First(&tt, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find ticket type: %w", err)
	}
	return &tt, nil
}

func (r *ticketTypeRepository) FindByEvent(ctx context.Context, eventID uuid.UUID, includeHidden bool) ([]models.TicketType, error) {
	var list []models.TicketType
	q := r.db.WithContext(ctx).Model(&models.TicketType{}).Where("event_id = ?", eventID)
	if !includeHidden {
		q = q.Where("is_hidden = ? AND is_active = ?", false, true)
	}
	if err := q.Order("sort_order asc, price_minor asc").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to list ticket types: %w", err)
	}
	return list, nil
}

func (r *ticketTypeRepository) Update(ctx context.Context, tt *models.TicketType) error {
	if err := r.db.WithContext(ctx).Save(tt).Error; err != nil {
		return fmt.Errorf("failed to update ticket type: %w", err)
	}
	return nil
}

func (r *ticketTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.TicketType{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete ticket type: %w", err)
	}
	return nil
}

// ReserveQuantity atomically increments quantity_reserved if there is enough stock.
// Uses a conditional UPDATE to prevent overselling under concurrency.
func (r *ticketTypeRepository) ReserveQuantity(ctx context.Context, ticketTypeID uuid.UUID, qty int) error {
	if qty <= 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Exec(`
		UPDATE ticket_types
		SET quantity_reserved = quantity_reserved + ?,
		    updated_at = NOW()
		WHERE id = ?
		  AND is_active = TRUE
		  AND (quantity_total - quantity_sold - quantity_reserved) >= ?
	`, qty, ticketTypeID, qty)
	if res.Error != nil {
		return fmt.Errorf("failed to reserve quantity: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("insufficient stock for ticket type %s", ticketTypeID)
	}
	return nil
}

// ReleaseReservation decrements quantity_reserved (used when an order fails/expires).
func (r *ticketTypeRepository) ReleaseReservation(ctx context.Context, ticketTypeID uuid.UUID, qty int) error {
	if qty <= 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Exec(`
		UPDATE ticket_types
		SET quantity_reserved = GREATEST(quantity_reserved - ?, 0),
		    updated_at = NOW()
		WHERE id = ?
	`, qty, ticketTypeID).Error; err != nil {
		return fmt.Errorf("failed to release reservation: %w", err)
	}
	return nil
}

// CommitReservation converts reserved -> sold (used when payment succeeds).
func (r *ticketTypeRepository) CommitReservation(ctx context.Context, ticketTypeID uuid.UUID, qty int) error {
	if qty <= 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Exec(`
		UPDATE ticket_types
		SET quantity_reserved = GREATEST(quantity_reserved - ?, 0),
		    quantity_sold = quantity_sold + ?,
		    updated_at = NOW()
		WHERE id = ?
	`, qty, qty, ticketTypeID)
	if res.Error != nil {
		return fmt.Errorf("failed to commit reservation: %w", res.Error)
	}
	return nil
}

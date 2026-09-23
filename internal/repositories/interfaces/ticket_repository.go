package interfaces

import (
	"context"
	"time"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type TicketTypeRepository interface {
	Create(ctx context.Context, tt *models.TicketType) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.TicketType, error)
	FindByEvent(ctx context.Context, eventID uuid.UUID, includeHidden bool) ([]models.TicketType, error)
	Update(ctx context.Context, tt *models.TicketType) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Atomic reservations for purchase flow
	ReserveQuantity(ctx context.Context, ticketTypeID uuid.UUID, qty int) error
	ReleaseReservation(ctx context.Context, ticketTypeID uuid.UUID, qty int) error
	CommitReservation(ctx context.Context, ticketTypeID uuid.UUID, qty int) error
}

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
	FindByReference(ctx context.Context, reference string) (*models.Order, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*models.Order, error)
	FindWithItems(ctx context.Context, id uuid.UUID) (*models.Order, error)
	FindWithTickets(ctx context.Context, id uuid.UUID) (*models.Order, error)
	Update(ctx context.Context, order *models.Order) error
	ListByUser(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Order, int64, error)
	ListByEvent(ctx context.Context, eventID uuid.UUID, page, perPage int) ([]models.Order, int64, error)
	FindExpiredPending(ctx context.Context, before time.Time, limit int) ([]models.Order, error)
}

type OrderItemRepository interface {
	CreateMany(ctx context.Context, items []models.OrderItem) error
	FindByOrder(ctx context.Context, orderID uuid.UUID) ([]models.OrderItem, error)
}

type TicketRepository interface {
	CreateMany(ctx context.Context, tickets []models.Ticket) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Ticket, error)
	FindByCode(ctx context.Context, code string) (*models.Ticket, error)
	FindByOrder(ctx context.Context, orderID uuid.UUID) ([]models.Ticket, error)
	FindByUser(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.Ticket, int64, error)
	FindByEvent(ctx context.Context, eventID uuid.UUID, page, perPage int) ([]models.Ticket, int64, error)
	Update(ctx context.Context, ticket *models.Ticket) error
	UpdateMany(ctx context.Context, tickets []models.Ticket) error
	CountByEvent(ctx context.Context, eventID uuid.UUID) (int64, error)
	CountUsedByEvent(ctx context.Context, eventID uuid.UUID) (int64, error)
}

type PromoCodeRepository interface {
	Create(ctx context.Context, pc *models.PromoCode) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.PromoCode, error)
	FindByCode(ctx context.Context, code string) (*models.PromoCode, error)
	FindByEvent(ctx context.Context, eventID uuid.UUID) ([]models.PromoCode, error)
	Update(ctx context.Context, pc *models.PromoCode) error
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementUsage(ctx context.Context, id uuid.UUID, delta int) error
	CountUserRedemptions(ctx context.Context, promoID, userID uuid.UUID) (int64, error)
}

type PromoCodeRedemptionRepository interface {
	Create(ctx context.Context, r *models.PromoCodeRedemption) error
	FindByOrder(ctx context.Context, orderID uuid.UUID) (*models.PromoCodeRedemption, error)
}

type CheckInRepository interface {
	Create(ctx context.Context, ci *models.CheckIn) error
	FindByTicket(ctx context.Context, ticketID uuid.UUID) (*models.CheckIn, error)
	FindByEvent(ctx context.Context, eventID uuid.UUID, page, perPage int) ([]models.CheckIn, int64, error)
	CountByEvent(ctx context.Context, eventID uuid.UUID) (int64, error)
}

type RefundRepository interface {
	Create(ctx context.Context, r *models.Refund) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Refund, error)
	FindByOrder(ctx context.Context, orderID uuid.UUID) ([]models.Refund, error)
	Update(ctx context.Context, r *models.Refund) error
	ListByStatus(ctx context.Context, status string, page, perPage int) ([]models.Refund, int64, error)
}

type LedgerRepository interface {
	Create(ctx context.Context, entry *models.LedgerEntry) error
	CreateMany(ctx context.Context, entries []models.LedgerEntry) error
	SumAccount(ctx context.Context, account models.LedgerAccount, ownerID *uuid.UUID, currency string) (creditMinor, debitMinor int64, err error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, page, perPage int) ([]models.LedgerEntry, int64, error)
	ListByReference(ctx context.Context, reference string) ([]models.LedgerEntry, error)
}

type PaymentTransactionRepository interface {
	Create(ctx context.Context, txn *models.PaymentTransaction) error
	FindByReference(ctx context.Context, reference string) (*models.PaymentTransaction, error)
	FindByOrder(ctx context.Context, orderID uuid.UUID) ([]models.PaymentTransaction, error)
	Update(ctx context.Context, txn *models.PaymentTransaction) error
}

type TicketTransferRepository interface {
	Create(ctx context.Context, t *models.TicketTransfer) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.TicketTransfer, error)
	FindByToken(ctx context.Context, token string) (*models.TicketTransfer, error)
	FindByTicket(ctx context.Context, ticketID uuid.UUID) ([]models.TicketTransfer, error)
	FindByUser(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.TicketTransfer, int64, error)
	Update(ctx context.Context, t *models.TicketTransfer) error
}

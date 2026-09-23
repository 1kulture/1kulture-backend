package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type LedgerService interface {
	// RecordOrderPaid is called when a paid order is confirmed.
	// It creates ledger entries crediting escrow (owed to organizer) and fee (platform commission).
	RecordOrderPaid(ctx context.Context, order *models.Order) error

	// ReleaseEscrow moves held funds from escrow to the organizer's available balance.
	ReleaseEscrow(ctx context.Context, order *models.Order) error

	// RecordRefund reverses funds. If escrow was not yet released, it debits escrow;
	// otherwise it debits the organizer's available balance.
	RecordRefund(ctx context.Context, order *models.Order, refund *models.Refund) error

	// RecordPayout records a payout to an organizer's bank account.
	RecordPayout(ctx context.Context, organizerID uuid.UUID, amountMinor int64, currency, reference string) error

	// GetBalance returns escrow-held, available, and paid-out balances for an owner.
	GetBalance(ctx context.Context, ownerID uuid.UUID, currency string) (*responses.BalanceResponse, error)

	// ListEntries returns paginated ledger entries for an owner.
	ListEntries(ctx context.Context, ownerID uuid.UUID, page, perPage int) ([]responses.LedgerEntryResponse, int64, error)
}

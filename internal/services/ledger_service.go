package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type ledgerService struct {
	ledgerRepo   repoInterfaces.LedgerRepository
	orderRepo    repoInterfaces.OrderRepository
	settingRepo  repoInterfaces.SettingRepository
	auditLogRepo repoInterfaces.AuditLogRepository
}

func NewLedgerService(
	ledgerRepo repoInterfaces.LedgerRepository,
	orderRepo repoInterfaces.OrderRepository,
	settingRepo repoInterfaces.SettingRepository,
	auditLogRepo repoInterfaces.AuditLogRepository,
) serviceInterfaces.LedgerService {
	return &ledgerService{
		ledgerRepo:   ledgerRepo,
		orderRepo:    orderRepo,
		settingRepo:  settingRepo,
		auditLogRepo: auditLogRepo,
	}
}

// ---------- Order paid → escrow + fee ----------

func (s *ledgerService) RecordOrderPaid(ctx context.Context, order *models.Order) error {
	if order.Status != models.OrderStatusPaid {
		return fmt.Errorf("cannot record ledger for non-paid order")
	}
	if order.TotalMinor <= 0 {
		// Free order — no money movement.
		return nil
	}

	ref := "order:" + order.Reference
	entries := []models.LedgerEntry{
		{
			Reference:         ref,
			Account:           models.LedgerAccountPlatform,
			EntryType:         models.LedgerEntryCredit,
			AmountMinor:       order.TotalMinor,
			Currency:          order.Currency,
			RelatedEntityType: "order",
			RelatedEntityID:   order.ID,
			Description:       fmt.Sprintf("Payment received for order %s", order.Reference),
		},
		{
			Reference:         ref,
			Account:           models.LedgerAccountEscrow,
			EntryType:         models.LedgerEntryCredit,
			OwnerID:           &order.Event.OrganizerID,
			AmountMinor:       order.OrganizerNetMinor,
			Currency:          order.Currency,
			RelatedEntityType: "order",
			RelatedEntityID:   order.ID,
			Description:       "Held in escrow pending release",
		},
		{
			Reference:         ref,
			Account:           models.LedgerAccountFee,
			EntryType:         models.LedgerEntryCredit,
			AmountMinor:       order.CommissionMinor,
			Currency:          order.Currency,
			RelatedEntityType: "order",
			RelatedEntityID:   order.ID,
			Description:       fmt.Sprintf("Platform commission (%d bps)", order.CommissionRateBps),
		},
	}

	if err := s.ledgerRepo.CreateMany(ctx, entries); err != nil {
		return err
	}

	s.audit(ctx, &order.UserID, "LEDGER_ORDER_PAID", ref)
	return nil
}

// ---------- Escrow release ----------

func (s *ledgerService) ReleaseEscrow(ctx context.Context, order *models.Order) error {
	if order.EscrowStatus == models.EscrowStatusReleased {
		return nil // idempotent
	}
	if order.OrganizerNetMinor <= 0 {
		return nil
	}

	ref := "escrow_release:" + order.Reference
	entries := []models.LedgerEntry{
		{
			Reference:         ref,
			Account:           models.LedgerAccountEscrow,
			EntryType:         models.LedgerEntryDebit,
			OwnerID:           &order.Event.OrganizerID,
			AmountMinor:       order.OrganizerNetMinor,
			Currency:          order.Currency,
			RelatedEntityType: "order",
			RelatedEntityID:   order.ID,
			Description:       "Escrow released to organizer",
		},
		{
			Reference:         ref,
			Account:           models.LedgerAccountOrganizer,
			EntryType:         models.LedgerEntryCredit,
			OwnerID:           &order.Event.OrganizerID,
			AmountMinor:       order.OrganizerNetMinor,
			Currency:          order.Currency,
			RelatedEntityType: "order",
			RelatedEntityID:   order.ID,
			Description:       "Funds available for payout",
		},
	}

	if err := s.ledgerRepo.CreateMany(ctx, entries); err != nil {
		return err
	}

	now := time.Now().UTC()
	order.EscrowStatus = models.EscrowStatusReleased
	order.EscrowReleasedAt = &now
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return fmt.Errorf("failed to update order escrow status: %w", err)
	}

	s.audit(ctx, nil, "LEDGER_ESCROW_RELEASED", ref)
	return nil
}

// ---------- Refund ----------

func (s *ledgerService) RecordRefund(ctx context.Context, order *models.Order, refund *models.Refund) error {
	if refund.AmountMinor <= 0 {
		return nil
	}

	ref := "refund:" + refund.ID.String()
	entries := []models.LedgerEntry{}

	if order.EscrowStatus == models.EscrowStatusReleased {
		// Escrow already released. Debit the organizer's available balance (may go negative — the platform absorbs it or chases it).
		entries = append(entries,
			models.LedgerEntry{
				Reference:         ref,
				Account:           models.LedgerAccountOrganizer,
				EntryType:         models.LedgerEntryDebit,
				OwnerID:           &order.Event.OrganizerID,
				AmountMinor:       refund.AmountMinor,
				Currency:          refund.Currency,
				RelatedEntityType: "refund",
				RelatedEntityID:   refund.ID,
				Description:       "Refund debited from organizer balance (post-release)",
			},
			models.LedgerEntry{
				Reference:         ref,
				Account:           models.LedgerAccountRefund,
				EntryType:         models.LedgerEntryCredit,
				AmountMinor:       refund.AmountMinor,
				Currency:          refund.Currency,
				RelatedEntityType: "refund",
				RelatedEntityID:   refund.ID,
				Description:       "Refund processed to buyer",
			},
		)
		refund.DebitedFromOrganizerBalance = true
	} else {
		// Escrow still held. Debit escrow and refund.
		entries = append(entries,
			models.LedgerEntry{
				Reference:         ref,
				Account:           models.LedgerAccountEscrow,
				EntryType:         models.LedgerEntryDebit,
				OwnerID:           &order.Event.OrganizerID,
				AmountMinor:       refund.AmountMinor,
				Currency:          refund.Currency,
				RelatedEntityType: "refund",
				RelatedEntityID:   refund.ID,
				Description:       "Refund reversed from escrow",
			},
			models.LedgerEntry{
				Reference:         ref,
				Account:           models.LedgerAccountRefund,
				EntryType:         models.LedgerEntryCredit,
				AmountMinor:       refund.AmountMinor,
				Currency:          refund.Currency,
				RelatedEntityType: "refund",
				RelatedEntityID:   refund.ID,
				Description:       "Refund processed to buyer",
			},
		)
	}

	if err := s.ledgerRepo.CreateMany(ctx, entries); err != nil {
		return err
	}

	s.audit(ctx, &refund.UserID, "LEDGER_REFUND_RECORDED", ref)
	return nil
}

// ---------- Payout ----------

func (s *ledgerService) RecordPayout(ctx context.Context, organizerID uuid.UUID, amountMinor int64, currency, reference string) error {
	if amountMinor <= 0 {
		return fmt.Errorf("payout amount must be positive")
	}

	entries := []models.LedgerEntry{
		{
			Reference:         reference,
			Account:           models.LedgerAccountOrganizer,
			EntryType:         models.LedgerEntryDebit,
			OwnerID:           &organizerID,
			AmountMinor:       amountMinor,
			Currency:          currency,
			RelatedEntityType: "payout",
			RelatedEntityID:   organizerID,
			Description:       "Payout debited from available balance",
		},
		{
			Reference:         reference,
			Account:           models.LedgerAccountPayout,
			EntryType:         models.LedgerEntryCredit,
			OwnerID:           &organizerID,
			AmountMinor:       amountMinor,
			Currency:          currency,
			RelatedEntityType: "payout",
			RelatedEntityID:   organizerID,
			Description:       "Payout sent to organizer bank account",
		},
	}
	if err := s.ledgerRepo.CreateMany(ctx, entries); err != nil {
		return err
	}
	s.audit(ctx, nil, "LEDGER_PAYOUT_RECORDED", reference)
	return nil
}

// ---------- Balances & listing ----------

func (s *ledgerService) GetBalance(ctx context.Context, ownerID uuid.UUID, currency string) (*responses.BalanceResponse, error) {
	// Escrow held: credits - debits on escrow account where owner_id = organizer
	escCr, escDb, err := s.ledgerRepo.SumAccount(ctx, models.LedgerAccountEscrow, &ownerID, currency)
	if err != nil {
		return nil, err
	}
	held := escCr - escDb

	// Available: credits - debits on organizer account
	orgCr, orgDb, err := s.ledgerRepo.SumAccount(ctx, models.LedgerAccountOrganizer, &ownerID, currency)
	if err != nil {
		return nil, err
	}
	available := orgCr - orgDb

	// Paid out: credits on payout account
	poCr, poDb, err := s.ledgerRepo.SumAccount(ctx, models.LedgerAccountPayout, &ownerID, currency)
	if err != nil {
		return nil, err
	}
	paidOut := poCr - poDb

	return &responses.BalanceResponse{
		OwnerID:        ownerID,
		Currency:       currency,
		HeldMinor:      held,
		AvailableMinor: available,
		PaidOutMinor:   paidOut,
	}, nil
}

func (s *ledgerService) ListEntries(ctx context.Context, ownerID uuid.UUID, page, perPage int) ([]responses.LedgerEntryResponse, int64, error) {
	list, total, err := s.ledgerRepo.ListByOwner(ctx, ownerID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.LedgerEntryResponse, 0, len(list))
	for i := range list {
		e := &list[i]
		out = append(out, responses.LedgerEntryResponse{
			ID:                e.ID,
			Reference:         e.Reference,
			Account:           string(e.Account),
			EntryType:         string(e.EntryType),
			OwnerID:           e.OwnerID,
			AmountMinor:       e.AmountMinor,
			Currency:          e.Currency,
			RelatedEntityType: e.RelatedEntityType,
			RelatedEntityID:   e.RelatedEntityID,
			Description:       e.Description,
			CreatedAt:         e.CreatedAt,
		})
	}
	return out, total, nil
}

func (s *ledgerService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "ledger",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log: ", err)
	}
}

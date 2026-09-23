package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/payments"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/idgen"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type refundService struct {
	refundRepo   repoInterfaces.RefundRepository
	orderRepo    repoInterfaces.OrderRepository
	ticketRepo   repoInterfaces.TicketRepository
	eventRepo    repoInterfaces.EventRepository
	coOrgRepo    repoInterfaces.EventCoOrganizerRepository
	ledgerSvc    serviceInterfaces.LedgerService
	provider     payments.PaymentProvider
	auditLogRepo repoInterfaces.AuditLogRepository
}

func NewRefundService(
	refundRepo repoInterfaces.RefundRepository,
	orderRepo repoInterfaces.OrderRepository,
	ticketRepo repoInterfaces.TicketRepository,
	eventRepo repoInterfaces.EventRepository,
	coOrgRepo repoInterfaces.EventCoOrganizerRepository,
	ledgerSvc serviceInterfaces.LedgerService,
	provider payments.PaymentProvider,
	auditLogRepo repoInterfaces.AuditLogRepository,
) serviceInterfaces.RefundService {
	return &refundService{
		refundRepo:   refundRepo,
		orderRepo:    orderRepo,
		ticketRepo:   ticketRepo,
		eventRepo:    eventRepo,
		coOrgRepo:    coOrgRepo,
		ledgerSvc:    ledgerSvc,
		provider:     provider,
		auditLogRepo: auditLogRepo,
	}
}

// ---------- RequestRefund ----------

func (s *refundService) RequestRefund(ctx context.Context, actorID, orderID uuid.UUID, req *requests.RefundRequestPayload) (*responses.RefundResponse, error) {
	order, err := s.orderRepo.FindWithItems(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.UserID != actorID {
		return nil, fmt.Errorf("forbidden")
	}
	if order.Status != models.OrderStatusPaid {
		return nil, fmt.Errorf("only paid orders can be refunded")
	}
	if order.PaidAt == nil {
		return nil, fmt.Errorf("order has no payment timestamp")
	}

	// Validate refund window (event's refund_policy_days)
	event, err := s.eventRepo.FindByID(ctx, order.EventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	days := event.RefundPolicyDays
	if days > 0 {
		deadline := event.StartAt.AddDate(0, 0, -days)
		if time.Now().UTC().After(deadline) {
			return nil, fmt.Errorf("refund window has closed for this event")
		}
	}

	// Load tickets and validate ownership + status
	allTickets, err := s.ticketRepo.FindByOrder(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	ticketMap := make(map[uuid.UUID]*models.Ticket, len(allTickets))
	for i := range allTickets {
		ticketMap[allTickets[i].ID] = &allTickets[i]
	}

	var refundAmount int64
	var refundTicketIDs []uuid.UUID
	for _, idStr := range req.TicketIDs {
		tid, err := uuid.Parse(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid ticket id: %s", idStr)
		}
		t, ok := ticketMap[tid]
		if !ok {
			return nil, fmt.Errorf("ticket %s is not part of this order", tid)
		}
		if t.UserID != actorID {
			return nil, fmt.Errorf("you do not own ticket %s", tid)
		}
		if t.Status == models.TicketStatusUsed {
			return nil, fmt.Errorf("ticket %s has already been used", t.Code)
		}
		if t.Status == models.TicketStatusRefunded || t.Status == models.TicketStatusCancelled {
			return nil, fmt.Errorf("ticket %s is not eligible for refund", t.Code)
		}
		refundTicketIDs = append(refundTicketIDs, tid)
		// Refund per-ticket price = order item unit price. We look it up by order_item.
		for _, item := range order.Items {
			if item.ID == t.OrderItemID {
				refundAmount += item.UnitPriceMinor
				break
			}
		}
	}

	if refundAmount <= 0 {
		return nil, fmt.Errorf("refund amount would be zero")
	}

	idsJSON, _ := jsonMarshal(refundTicketIDs)

	refund := &models.Refund{
		OrderID:       order.ID,
		UserID:        actorID,
		TicketIDs:     datatypes.JSON(idsJSON),
		AmountMinor:   refundAmount,
		Currency:      order.Currency,
		Reason:        req.Reason,
		Status:        models.RefundStatusRequested,
		RequestedByID: actorID,
	}
	if err := s.refundRepo.Create(ctx, refund); err != nil {
		return nil, err
	}

	s.audit(ctx, &actorID, "REFUND_REQUESTED", refund.ID.String())
	return toRefundResponse(refund), nil
}

// ---------- DecideRefund ----------

func (s *refundService) DecideRefund(ctx context.Context, actorID, refundID uuid.UUID, req *requests.RefundDecisionRequest) (*responses.RefundResponse, error) {
	refund, err := s.refundRepo.FindByID(ctx, refundID)
	if err != nil {
		return nil, err
	}
	if refund == nil {
		return nil, fmt.Errorf("refund not found")
	}
	if refund.Status != models.RefundStatusRequested {
		return nil, fmt.Errorf("refund has already been reviewed")
	}

	order, err := s.orderRepo.FindWithItems(ctx, refund.OrderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	event, err := s.eventRepo.FindByID(ctx, order.EventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}

	// Authorization: organizer, co-organizer, or admin
	isOrganizer := event.OrganizerID == actorID
	isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, event.ID, actorID)
	if !isOrganizer && !isCo {
		return nil, fmt.Errorf("forbidden")
	}

	now := time.Now().UTC()
	refund.ReviewedByID = &actorID
	refund.ReviewedAt = &now
	refund.ReviewNote = req.ReviewNote

	if req.Status == "rejected" {
		refund.Status = models.RefundStatusRejected
		if err := s.refundRepo.Update(ctx, refund); err != nil {
			return nil, err
		}
		s.audit(ctx, &actorID, "REFUND_REJECTED", refund.ID.String())
		return toRefundResponse(refund), nil
	}

	// Approved → attempt provider refund, then update tickets + ledger
	refund.Status = models.RefundStatusApproved
	if err := s.refundRepo.Update(ctx, refund); err != nil {
		return nil, err
	}

	providerRef := ""
	if order.PaymentReference != "" {
		res, err := s.provider.Refund(ctx, payments.RefundRequest{
			TransactionReference: order.PaymentReference,
			AmountMinor:          refund.AmountMinor,
			Currency:             order.Currency,
			Reason:               refund.Reason,
			RefundReference:      idgen.RefundReference(),
		})
		if err != nil {
			logger.Error("Provider refund failed: ", err)
			refund.Status = models.RefundStatusFailed
			refund.ReviewNote += " | provider error: " + err.Error()
			_ = s.refundRepo.Update(ctx, refund)
			return toRefundResponse(refund), fmt.Errorf("provider refund failed: %w", err)
		}
		providerRef = res.ProviderRef
	}

	// Cancel refunded tickets
	var ids []uuid.UUID
	_ = jsonUnmarshal(refund.TicketIDs, &ids)
	for _, tid := range ids {
		t, err := s.ticketRepo.FindByID(ctx, tid)
		if err != nil || t == nil {
			continue
		}
		t.Status = models.TicketStatusRefunded
		if err := s.ticketRepo.Update(ctx, t); err != nil {
			logger.Error("Failed to mark ticket refunded: ", err)
		}
	}

	// Ledger entry
	if err := s.ledgerSvc.RecordRefund(ctx, order, refund); err != nil {
		logger.Error("Failed to record refund ledger: ", err)
	}

	processedAt := time.Now().UTC()
	refund.Status = models.RefundStatusProcessed
	refund.ProviderRefundReference = providerRef
	refund.ProcessedAt = &processedAt
	if err := s.refundRepo.Update(ctx, refund); err != nil {
		return nil, err
	}

	// Update order status
	totalRefunded := int64(0)
	existing, _ := s.refundRepo.FindByOrder(ctx, order.ID)
	for _, r := range existing {
		if r.Status == models.RefundStatusProcessed {
			totalRefunded += r.AmountMinor
		}
	}
	if totalRefunded >= order.TotalMinor {
		order.Status = models.OrderStatusRefunded
		order.EscrowStatus = models.EscrowStatusRefunded
	} else if totalRefunded > 0 {
		order.Status = models.OrderStatusPartiallyRefunded
		order.EscrowStatus = models.EscrowStatusPartiallyRefunded
	}
	if err := s.orderRepo.Update(ctx, order); err != nil {
		logger.Error("Failed to update order after refund: ", err)
	}

	s.audit(ctx, &actorID, "REFUND_PROCESSED", refund.ID.String())
	return toRefundResponse(refund), nil
}

// ---------- Reads ----------

func (s *refundService) GetRefund(ctx context.Context, viewerID, refundID uuid.UUID) (*responses.RefundResponse, error) {
	refund, err := s.refundRepo.FindByID(ctx, refundID)
	if err != nil {
		return nil, err
	}
	if refund == nil {
		return nil, fmt.Errorf("refund not found")
	}
	if refund.UserID == viewerID || refund.RequestedByID == viewerID {
		return toRefundResponse(refund), nil
	}
	order, _ := s.orderRepo.FindByID(ctx, refund.OrderID)
	if order != nil {
		event, _ := s.eventRepo.FindByID(ctx, order.EventID)
		if event != nil {
			if event.OrganizerID == viewerID {
				return toRefundResponse(refund), nil
			}
			isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, event.ID, viewerID)
			if isCo {
				return toRefundResponse(refund), nil
			}
		}
	}
	return nil, fmt.Errorf("forbidden")
}

func (s *refundService) ListByOrder(ctx context.Context, viewerID, orderID uuid.UUID) ([]responses.RefundResponse, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.UserID != viewerID {
		event, _ := s.eventRepo.FindByID(ctx, order.EventID)
		if event == nil || event.OrganizerID != viewerID {
			isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, order.EventID, viewerID)
			if !isCo {
				return nil, fmt.Errorf("forbidden")
			}
		}
	}
	list, err := s.refundRepo.FindByOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}
	out := make([]responses.RefundResponse, 0, len(list))
	for i := range list {
		out = append(out, *toRefundResponse(&list[i]))
	}
	return out, nil
}

func (s *refundService) ListByStatus(ctx context.Context, status string, page, perPage int) ([]responses.RefundResponse, int64, error) {
	list, total, err := s.refundRepo.ListByStatus(ctx, status, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.RefundResponse, 0, len(list))
	for i := range list {
		out = append(out, *toRefundResponse(&list[i]))
	}
	return out, total, nil
}

// ---------- helpers ----------

func (s *refundService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "refund",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log: ", err)
	}
}

func toRefundResponse(r *models.Refund) *responses.RefundResponse {
	var ids []uuid.UUID
	if len(r.TicketIDs) > 0 {
		_ = jsonUnmarshal(r.TicketIDs, &ids)
	}
	return &responses.RefundResponse{
		ID:                          r.ID,
		OrderID:                     r.OrderID,
		UserID:                      r.UserID,
		TicketIDs:                   ids,
		AmountMinor:                 r.AmountMinor,
		Currency:                    r.Currency,
		Reason:                      r.Reason,
		Status:                      string(r.Status),
		RequestedByID:               r.RequestedByID,
		ReviewedByID:                r.ReviewedByID,
		ReviewedAt:                  r.ReviewedAt,
		ReviewNote:                  r.ReviewNote,
		ProviderRefundReference:     r.ProviderRefundReference,
		ProcessedAt:                 r.ProcessedAt,
		DebitedFromOrganizerBalance: r.DebitedFromOrganizerBalance,
		CreatedAt:                   r.CreatedAt,
		UpdatedAt:                   r.UpdatedAt,
	}
}

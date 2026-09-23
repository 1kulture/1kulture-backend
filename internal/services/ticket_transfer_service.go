package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/idgen"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type ticketTransferService struct {
	transferRepo repoInterfaces.TicketTransferRepository
	ticketRepo   repoInterfaces.TicketRepository
	userRepo     repoInterfaces.UserRepository
	settingSvc   serviceInterfaces.SettingService
	auditLogRepo repoInterfaces.AuditLogRepository
}

func NewTicketTransferService(
	transferRepo repoInterfaces.TicketTransferRepository,
	ticketRepo repoInterfaces.TicketRepository,
	userRepo repoInterfaces.UserRepository,
	settingSvc serviceInterfaces.SettingService,
	auditLogRepo repoInterfaces.AuditLogRepository,
) serviceInterfaces.TicketTransferService {
	return &ticketTransferService{
		transferRepo: transferRepo,
		ticketRepo:   ticketRepo,
		userRepo:     userRepo,
		settingSvc:   settingSvc,
		auditLogRepo: auditLogRepo,
	}
}

func (s *ticketTransferService) InitiateTransfer(ctx context.Context, actorID, ticketID uuid.UUID, req *requests.InitiateTransferRequest) (*responses.TicketTransferResponse, error) {
	// Feature flag
	if !s.settingSvc.GetBool(ctx, models.SettingTicketTransferEnabled, true) {
		return nil, fmt.Errorf("ticket transfers are currently disabled")
	}

	ticket, err := s.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, fmt.Errorf("ticket not found")
	}
	if ticket.UserID != actorID {
		return nil, fmt.Errorf("forbidden")
	}
	if ticket.Status == models.TicketStatusUsed ||
		ticket.Status == models.TicketStatusRefunded ||
		ticket.Status == models.TicketStatusCancelled {
		return nil, fmt.Errorf("ticket cannot be transferred (status: %s)", ticket.Status)
	}

	// Check for existing pending transfer on this ticket
	existing, err := s.transferRepo.FindByTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	for _, t := range existing {
		if t.Status == models.TicketTransferPending && time.Now().Before(t.ExpiresAt) {
			return nil, fmt.Errorf("ticket already has a pending transfer")
		}
	}

	toEmail := strings.ToLower(strings.TrimSpace(req.ToEmail))

	// Prevent self-transfer
	if u, err := s.userRepo.FindByID(ctx, actorID); err == nil && u != nil && strings.EqualFold(u.Email, toEmail) {
		return nil, fmt.Errorf("cannot transfer a ticket to yourself")
	}

	// Try to resolve recipient user ID if they already have an account
	var toUserID *uuid.UUID
	if u, err := s.userRepo.FindByEmail(ctx, toEmail); err == nil && u != nil {
		toUserID = &u.ID
	}

	transfer := &models.TicketTransfer{
		TicketID:   ticketID,
		FromUserID: actorID,
		ToEmail:    toEmail,
		ToUserID:   toUserID,
		Message:    req.Message,
		Token:      idgen.TransferToken(),
		Status:     models.TicketTransferPending,
		ExpiresAt:  time.Now().UTC().AddDate(0, 0, 7), // 7-day window
	}
	if err := s.transferRepo.Create(ctx, transfer); err != nil {
		return nil, err
	}

	s.audit(ctx, &actorID, "TICKET_TRANSFER_INITIATED", transfer.ID.String())
	return toTransferResponse(transfer), nil
}

func (s *ticketTransferService) AcceptTransfer(ctx context.Context, actorID uuid.UUID, req *requests.AcceptTransferRequest) (*responses.TicketResponse, error) {
	transfer, err := s.transferRepo.FindByToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	if transfer == nil {
		return nil, fmt.Errorf("transfer not found")
	}
	if !transfer.IsPending() {
		return nil, fmt.Errorf("transfer is no longer valid")
	}

	user, err := s.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	if !strings.EqualFold(user.Email, transfer.ToEmail) {
		return nil, fmt.Errorf("this transfer was not intended for your email")
	}

	ticket, err := s.ticketRepo.FindByID(ctx, transfer.TicketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, fmt.Errorf("ticket not found")
	}
	if ticket.UserID != transfer.FromUserID {
		return nil, fmt.Errorf("ticket ownership changed; transfer is no longer valid")
	}

	now := time.Now().UTC()
	// Transfer ticket to new owner
	ticket.UserID = actorID
	ticket.TransferCount++
	ticket.HolderName = user.FirstName + " " + user.LastName
	ticket.HolderEmail = user.Email
	ticket.HolderPhone = user.PhoneNumber
	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	transfer.Status = models.TicketTransferAccepted
	transfer.ToUserID = &actorID
	transfer.AcceptedAt = &now
	if err := s.transferRepo.Update(ctx, transfer); err != nil {
		return nil, err
	}

	s.audit(ctx, &actorID, "TICKET_TRANSFER_ACCEPTED", transfer.ID.String())
	return toTicketResponse(ticket), nil
}

func (s *ticketTransferService) DeclineTransfer(ctx context.Context, actorID uuid.UUID, req *requests.DeclineTransferRequest) (*responses.TicketTransferResponse, error) {
	transfer, err := s.transferRepo.FindByToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	if transfer == nil {
		return nil, fmt.Errorf("transfer not found")
	}
	if !transfer.IsPending() {
		return nil, fmt.Errorf("transfer is no longer valid")
	}

	user, err := s.userRepo.FindByID(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if user == nil || !strings.EqualFold(user.Email, transfer.ToEmail) {
		return nil, fmt.Errorf("this transfer was not intended for your email")
	}

	now := time.Now().UTC()
	transfer.Status = models.TicketTransferDeclined
	transfer.DeclinedAt = &now
	if err := s.transferRepo.Update(ctx, transfer); err != nil {
		return nil, err
	}

	s.audit(ctx, &actorID, "TICKET_TRANSFER_DECLINED", transfer.ID.String())
	return toTransferResponse(transfer), nil
}

func (s *ticketTransferService) CancelTransfer(ctx context.Context, actorID, transferID uuid.UUID) (*responses.TicketTransferResponse, error) {
	transfer, err := s.transferRepo.FindByID(ctx, transferID)
	if err != nil {
		return nil, err
	}
	if transfer == nil {
		return nil, fmt.Errorf("transfer not found")
	}
	if transfer.FromUserID != actorID {
		return nil, fmt.Errorf("forbidden")
	}
	if !transfer.IsPending() {
		return nil, fmt.Errorf("transfer can no longer be cancelled")
	}

	now := time.Now().UTC()
	transfer.Status = models.TicketTransferCancelled
	transfer.CancelledAt = &now
	if err := s.transferRepo.Update(ctx, transfer); err != nil {
		return nil, err
	}

	s.audit(ctx, &actorID, "TICKET_TRANSFER_CANCELLED", transfer.ID.String())
	return toTransferResponse(transfer), nil
}

func (s *ticketTransferService) GetTransfer(ctx context.Context, viewerID, transferID uuid.UUID) (*responses.TicketTransferResponse, error) {
	transfer, err := s.transferRepo.FindByID(ctx, transferID)
	if err != nil {
		return nil, err
	}
	if transfer == nil {
		return nil, fmt.Errorf("transfer not found")
	}
	if transfer.FromUserID != viewerID {
		if transfer.ToUserID == nil || *transfer.ToUserID != viewerID {
			return nil, fmt.Errorf("forbidden")
		}
	}
	return toTransferResponse(transfer), nil
}

func (s *ticketTransferService) GetTransferByToken(ctx context.Context, token string) (*responses.TicketTransferResponse, error) {
	transfer, err := s.transferRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if transfer == nil {
		return nil, fmt.Errorf("transfer not found")
	}
	return toTransferResponse(transfer), nil
}

func (s *ticketTransferService) ListMyTransfers(ctx context.Context, userID uuid.UUID, page, perPage int) ([]responses.TicketTransferResponse, int64, error) {
	list, total, err := s.transferRepo.FindByUser(ctx, userID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.TicketTransferResponse, 0, len(list))
	for i := range list {
		out = append(out, *toTransferResponse(&list[i]))
	}
	return out, total, nil
}

func (s *ticketTransferService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "ticket_transfer",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log: ", err)
	}
}

func toTransferResponse(t *models.TicketTransfer) *responses.TicketTransferResponse {
	return &responses.TicketTransferResponse{
		ID:          t.ID,
		TicketID:    t.TicketID,
		FromUserID:  t.FromUserID,
		ToEmail:     t.ToEmail,
		ToUserID:    t.ToUserID,
		Message:     t.Message,
		Status:      string(t.Status),
		ExpiresAt:   t.ExpiresAt,
		AcceptedAt:  t.AcceptedAt,
		DeclinedAt:  t.DeclinedAt,
		CancelledAt: t.CancelledAt,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

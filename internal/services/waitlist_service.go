package services

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type waitlistService struct {
	waitlistRepo interfaces.WaitlistRepository
	auditLogRepo interfaces.AuditLogRepository
}

func NewWaitlistService(
	waitlistRepo interfaces.WaitlistRepository,
	auditLogRepo interfaces.AuditLogRepository,
) serviceInterfaces.WaitlistService {
	return &waitlistService{
		waitlistRepo: waitlistRepo,
		auditLogRepo: auditLogRepo,
	}
}

func (s *waitlistService) AddToWaitlist(ctx context.Context, req *requests.WaitlistRequest) error {
	entry := &models.WaitlistEntry{
		Email:    req.Email,
		Category: req.Category,
	}
	if err := s.waitlistRepo.Create(ctx, entry); err != nil {
		return err
	}

	// Optional audit log (no authenticated user, so UserID nil)
	auditLog := &models.AuditLog{
		Action:     "WAITLIST_ADD",
		Resource:   "waitlist",
		ResourceID: entry.ID.String(),
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
		logger.Error("Failed to create audit log for waitlist entry:", err)
		// don't fail the main operation
	}

	return nil
}

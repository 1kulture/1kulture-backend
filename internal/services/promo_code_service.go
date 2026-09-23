package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type promoCodeService struct {
	promoRepo      repoInterfaces.PromoCodeRepository
	redemptionRepo repoInterfaces.PromoCodeRedemptionRepository
	eventRepo      repoInterfaces.EventRepository
	coOrgRepo      repoInterfaces.EventCoOrganizerRepository
	auditLogRepo   repoInterfaces.AuditLogRepository
}

func NewPromoCodeService(
	promoRepo repoInterfaces.PromoCodeRepository,
	redemptionRepo repoInterfaces.PromoCodeRedemptionRepository,
	eventRepo repoInterfaces.EventRepository,
	coOrgRepo repoInterfaces.EventCoOrganizerRepository,
	auditLogRepo repoInterfaces.AuditLogRepository,
) serviceInterfaces.PromoCodeService {
	return &promoCodeService{
		promoRepo:      promoRepo,
		redemptionRepo: redemptionRepo,
		eventRepo:      eventRepo,
		coOrgRepo:      coOrgRepo,
		auditLogRepo:   auditLogRepo,
	}
}

func (s *promoCodeService) canManage(ctx context.Context, event *models.Event, userID uuid.UUID) bool {
	if event.OrganizerID == userID {
		return true
	}
	isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, event.ID, userID)
	return isCo
}

func (s *promoCodeService) Create(ctx context.Context, actorID, eventID uuid.UUID, req *requests.PromoCodeCreateRequest) (*responses.PromoCodeResponse, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	if !s.canManage(ctx, event, actorID) {
		return nil, fmt.Errorf("forbidden")
	}

	code := strings.ToUpper(strings.TrimSpace(req.Code))
	existing, _ := s.promoRepo.FindByCode(ctx, code)
	if existing != nil {
		return nil, fmt.Errorf("promo code '%s' already exists", code)
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	applicableIDs := []string{}
	if len(req.ApplicableTicketTypeIDs) > 0 {
		applicableIDs = req.ApplicableTicketTypeIDs
	}
	applicableJSON, _ := jsonMarshal(applicableIDs)

	eid := eventID
	pc := &models.PromoCode{
		EventID:                 &eid,
		Code:                    code,
		Type:                    models.PromoType(req.Type),
		ValueMinor:              req.ValueMinor,
		UsageLimit:              req.UsageLimit,
		PerUserLimit:            req.PerUserLimit,
		ValidFrom:               req.ValidFrom,
		ValidTo:                 req.ValidTo,
		MinOrderMinor:           req.MinOrderMinor,
		ApplicableTicketTypeIDs: datatypes.JSON(applicableJSON),
		IsActive:                isActive,
		CreatedByID:             actorID,
	}
	if err := s.promoRepo.Create(ctx, pc); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "PROMO_CODE_CREATED", pc.ID.String())
	return toPromoCodeResponse(pc), nil
}

func (s *promoCodeService) List(ctx context.Context, actorID, eventID uuid.UUID) ([]responses.PromoCodeResponse, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	if !s.canManage(ctx, event, actorID) {
		return nil, fmt.Errorf("forbidden")
	}
	list, err := s.promoRepo.FindByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	out := make([]responses.PromoCodeResponse, 0, len(list))
	for i := range list {
		out = append(out, *toPromoCodeResponse(&list[i]))
	}
	return out, nil
}

func (s *promoCodeService) Update(ctx context.Context, actorID, eventID, promoID uuid.UUID, req *requests.PromoCodeUpdateRequest) (*responses.PromoCodeResponse, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	if !s.canManage(ctx, event, actorID) {
		return nil, fmt.Errorf("forbidden")
	}
	pc, err := s.promoRepo.FindByID(ctx, promoID)
	if err != nil {
		return nil, err
	}
	if pc == nil || pc.EventID == nil || *pc.EventID != eventID {
		return nil, fmt.Errorf("promo code not found")
	}
	if req.UsageLimit != nil {
		pc.UsageLimit = *req.UsageLimit
	}
	if req.PerUserLimit != nil {
		pc.PerUserLimit = *req.PerUserLimit
	}
	if req.ValidFrom != nil {
		pc.ValidFrom = req.ValidFrom
	}
	if req.ValidTo != nil {
		pc.ValidTo = req.ValidTo
	}
	if req.MinOrderMinor != nil {
		pc.MinOrderMinor = *req.MinOrderMinor
	}
	if req.IsActive != nil {
		pc.IsActive = *req.IsActive
	}
	if err := s.promoRepo.Update(ctx, pc); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "PROMO_CODE_UPDATED", pc.ID.String())
	return toPromoCodeResponse(pc), nil
}

func (s *promoCodeService) Delete(ctx context.Context, actorID, eventID, promoID uuid.UUID) error {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return err
	}
	if event == nil {
		return fmt.Errorf("event not found")
	}
	if !s.canManage(ctx, event, actorID) {
		return fmt.Errorf("forbidden")
	}
	if err := s.promoRepo.Delete(ctx, promoID); err != nil {
		return err
	}
	s.audit(ctx, &actorID, "PROMO_CODE_DELETED", promoID.String())
	return nil
}

func (s *promoCodeService) Validate(ctx context.Context, userID uuid.UUID, req *requests.ValidatePromoRequest) (*responses.PromoValidationResponse, error) {
	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		return nil, fmt.Errorf("invalid event_id")
	}
	code := strings.ToUpper(strings.TrimSpace(req.PromoCode))

	pc, err := s.promoRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if pc == nil {
		return nil, fmt.Errorf("invalid promo code")
	}
	if pc.EventID != nil && *pc.EventID != eventID {
		return nil, fmt.Errorf("promo code is not valid for this event")
	}
	now := time.Now().UTC()
	if !pc.IsValid(now) {
		return nil, fmt.Errorf("promo code is not active")
	}
	if pc.MinOrderMinor > 0 && req.SubtotalMinor < pc.MinOrderMinor {
		return nil, fmt.Errorf("order subtotal does not meet the minimum for this promo code")
	}

	// Per-user limit
	if pc.PerUserLimit > 0 {
		count, err := s.promoRepo.CountUserRedemptions(ctx, pc.ID, userID)
		if err != nil {
			return nil, err
		}
		if int(count) >= pc.PerUserLimit {
			return nil, fmt.Errorf("you have reached the usage limit for this promo code")
		}
	}

	discount := calculateDiscount(pc, req.SubtotalMinor)
	if discount > req.SubtotalMinor {
		discount = req.SubtotalMinor
	}

	return &responses.PromoValidationResponse{
		Code:             pc.Code,
		DiscountMinor:    discount,
		NewSubtotalMinor: req.SubtotalMinor - discount,
		Message:          "Promo code applied",
	}, nil
}

// calculateDiscount is shared with OrderService.
func calculateDiscount(pc *models.PromoCode, subtotal int64) int64 {
	if subtotal <= 0 {
		return 0
	}
	switch pc.Type {
	case models.PromoTypePercentage:
		// ValueMinor holds basis points (500 = 5%)
		return subtotal * pc.ValueMinor / 10000
	case models.PromoTypeFixed:
		return pc.ValueMinor
	}
	return 0
}

func (s *promoCodeService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "promo_code",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log: ", err)
	}
}

func toPromoCodeResponse(pc *models.PromoCode) *responses.PromoCodeResponse {
	var applicableIDs []string
	if len(pc.ApplicableTicketTypeIDs) > 0 && string(pc.ApplicableTicketTypeIDs) != "null" {
		_ = jsonUnmarshal(pc.ApplicableTicketTypeIDs, &applicableIDs)
	}
	return &responses.PromoCodeResponse{
		ID:                      pc.ID,
		EventID:                 pc.EventID,
		Code:                    pc.Code,
		Type:                    string(pc.Type),
		ValueMinor:              pc.ValueMinor,
		UsageLimit:              pc.UsageLimit,
		UsageCount:              pc.UsageCount,
		PerUserLimit:            pc.PerUserLimit,
		ValidFrom:               pc.ValidFrom,
		ValidTo:                 pc.ValidTo,
		MinOrderMinor:           pc.MinOrderMinor,
		ApplicableTicketTypeIDs: applicableIDs,
		IsActive:                pc.IsActive,
		CreatedAt:               pc.CreatedAt,
		UpdatedAt:               pc.UpdatedAt,
	}
}

package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type affiliateCodeService struct {
	affiliateRepo repoInterfaces.AffiliateCodeRepository
	promoRepo     repoInterfaces.PromoCodeRepository
	requestRepo   repoInterfaces.PartnershipRequestRepository
	eventRepo     repoInterfaces.EventRepository
	coOrgRepo     repoInterfaces.EventCoOrganizerRepository
	auditLogRepo  repoInterfaces.AuditLogRepository
}

func NewAffiliateCodeService(
	affiliateRepo repoInterfaces.AffiliateCodeRepository,
	promoRepo repoInterfaces.PromoCodeRepository,
	requestRepo repoInterfaces.PartnershipRequestRepository,
	eventRepo repoInterfaces.EventRepository,
	coOrgRepo repoInterfaces.EventCoOrganizerRepository,
	auditLogRepo repoInterfaces.AuditLogRepository,
) serviceInterfaces.AffiliateCodeService {
	return &affiliateCodeService{
		affiliateRepo: affiliateRepo,
		promoRepo:     promoRepo,
		requestRepo:   requestRepo,
		eventRepo:     eventRepo,
		coOrgRepo:     coOrgRepo,
		auditLogRepo:  auditLogRepo,
	}
}

func (s *affiliateCodeService) Create(ctx context.Context, actorID, partnershipID uuid.UUID, req *requests.CreateAffiliateCodeRequest) (*responses.AffiliateCodeResponse, error) {
	pr, err := s.requestRepo.FindByID(ctx, partnershipID)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, fmt.Errorf("partnership not found")
	}

	// Only the organizer (or co-org) creates the promo code tied to a partnership.
	if pr.OrganizerID != actorID {
		isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, pr.EventID, actorID)
		if !isCo {
			return nil, fmt.Errorf("forbidden")
		}
	}

	if !pr.IsActiveState() {
		return nil, fmt.Errorf("affiliate codes can only be created for accepted or active partnerships")
	}

	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if code == "" {
		return nil, fmt.Errorf("code is required")
	}

	// Check affiliate code uniqueness
	if existing, _ := s.affiliateRepo.FindByCode(ctx, code); existing != nil {
		return nil, fmt.Errorf("affiliate code '%s' already exists", code)
	}
	// Also check against promo codes
	if existing, _ := s.promoRepo.FindByCode(ctx, code); existing != nil {
		return nil, fmt.Errorf("code '%s' already exists as a promo code", code)
	}

	// Create the underlying promo code tied to the event
	eid := pr.EventID
	promo := &models.PromoCode{
		EventID:                 &eid,
		Code:                    code,
		Type:                    models.PromoType(req.Type),
		ValueMinor:              req.ValueMinor,
		UsageLimit:              req.UsageLimit,
		PerUserLimit:            req.PerUserLimit,
		IsActive:                true,
		CreatedByID:             actorID,
		ApplicableTicketTypeIDs: datatypes.JSON([]byte("[]")),
	}
	if err := s.promoRepo.Create(ctx, promo); err != nil {
		return nil, err
	}

	affiliate := &models.AffiliateCode{
		PartnershipID: partnershipID,
		PromoCodeID:   promo.ID,
		Code:          code,
	}
	if err := s.affiliateRepo.Create(ctx, affiliate); err != nil {
		return nil, err
	}

	s.audit(ctx, &actorID, "AFFILIATE_CODE_CREATED", affiliate.ID.String())
	return toAffiliateCodeResponse(affiliate), nil
}

func (s *affiliateCodeService) List(ctx context.Context, viewerID, partnershipID uuid.UUID) ([]responses.AffiliateCodeResponse, error) {
	pr, err := s.requestRepo.FindByID(ctx, partnershipID)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, fmt.Errorf("partnership not found")
	}
	// Brand or organizer can view
	if pr.BrandUserID != viewerID && pr.OrganizerID != viewerID {
		isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, pr.EventID, viewerID)
		if !isCo {
			return nil, fmt.Errorf("forbidden")
		}
	}
	list, err := s.affiliateRepo.FindByPartnership(ctx, partnershipID)
	if err != nil {
		return nil, err
	}
	out := make([]responses.AffiliateCodeResponse, 0, len(list))
	for i := range list {
		out = append(out, *toAffiliateCodeResponse(&list[i]))
	}
	return out, nil
}

func (s *affiliateCodeService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "affiliate_code",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log: ", err)
	}
}

func toAffiliateCodeResponse(a *models.AffiliateCode) *responses.AffiliateCodeResponse {
	return &responses.AffiliateCodeResponse{
		ID:              a.ID,
		PartnershipID:   a.PartnershipID,
		PromoCodeID:     a.PromoCodeID,
		Code:            a.Code,
		Clicks:          a.Clicks,
		Conversions:     a.Conversions,
		RevenueMinor:    a.RevenueMinor,
		CommissionMinor: a.CommissionMinor,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
}

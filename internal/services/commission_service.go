package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type commissionService struct {
	tierRepo     interfaces.CommissionTierRepository
	overrideRepo interfaces.OrganizerCommissionOverrideRepository
	settingRepo  interfaces.SettingRepository
	auditLogRepo interfaces.AuditLogRepository
}

func NewCommissionService(
	tierRepo interfaces.CommissionTierRepository,
	overrideRepo interfaces.OrganizerCommissionOverrideRepository,
	settingRepo interfaces.SettingRepository,
	auditLogRepo interfaces.AuditLogRepository,
) serviceInterfaces.CommissionService {
	return &commissionService{
		tierRepo:     tierRepo,
		overrideRepo: overrideRepo,
		settingRepo:  settingRepo,
		auditLogRepo: auditLogRepo,
	}
}

func (s *commissionService) CreateTier(ctx context.Context, req *requests.CreateCommissionTierRequest) (*responses.CommissionTierResponse, error) {
	t := &models.CommissionTier{
		Name:            req.Name,
		MinRevenueMinor: req.MinRevenueMinor,
		MaxRevenueMinor: req.MaxRevenueMinor,
		RateBps:         req.RateBps,
		Priority:        req.Priority,
		IsActive:        true,
	}
	if req.IsActive != nil {
		t.IsActive = *req.IsActive
	}
	if err := s.tierRepo.Create(ctx, t); err != nil {
		return nil, err
	}
	return toCommissionTierResponse(t), nil
}

func (s *commissionService) ListTiers(ctx context.Context, onlyActive bool) ([]responses.CommissionTierResponse, error) {
	list, err := s.tierRepo.FindAll(ctx, onlyActive)
	if err != nil {
		return nil, err
	}
	out := make([]responses.CommissionTierResponse, 0, len(list))
	for i := range list {
		out = append(out, *toCommissionTierResponse(&list[i]))
	}
	return out, nil
}

func (s *commissionService) UpdateTier(ctx context.Context, id uuid.UUID, req *requests.UpdateCommissionTierRequest) (*responses.CommissionTierResponse, error) {
	t, err := s.tierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, fmt.Errorf("commission tier not found")
	}
	if req.Name != "" {
		t.Name = req.Name
	}
	if req.MinRevenueMinor != nil {
		t.MinRevenueMinor = *req.MinRevenueMinor
	}
	if req.MaxRevenueMinor != nil {
		t.MaxRevenueMinor = req.MaxRevenueMinor
	}
	if req.RateBps != nil {
		t.RateBps = *req.RateBps
	}
	if req.Priority != nil {
		t.Priority = *req.Priority
	}
	if req.IsActive != nil {
		t.IsActive = *req.IsActive
	}
	if err := s.tierRepo.Update(ctx, t); err != nil {
		return nil, err
	}
	return toCommissionTierResponse(t), nil
}

func (s *commissionService) DeleteTier(ctx context.Context, id uuid.UUID) error {
	return s.tierRepo.Delete(ctx, id)
}

func (s *commissionService) SetOverride(ctx context.Context, actorID uuid.UUID, req *requests.SetOrganizerCommissionOverrideRequest) (*responses.OrganizerCommissionOverrideResponse, error) {
	organizerID, err := uuid.Parse(req.OrganizerID)
	if err != nil {
		return nil, fmt.Errorf("invalid organizer id")
	}
	o := &models.OrganizerCommissionOverride{
		OrganizerID: organizerID,
		RateBps:     req.RateBps,
		Reason:      req.Reason,
		SetBy:       &actorID,
	}
	if err := s.overrideRepo.Upsert(ctx, o); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "COMMISSION_OVERRIDE_SET", organizerID.String())
	return toOverrideResponse(o), nil
}

func (s *commissionService) GetOverride(ctx context.Context, organizerID uuid.UUID) (*responses.OrganizerCommissionOverrideResponse, error) {
	o, err := s.overrideRepo.FindByOrganizer(ctx, organizerID)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, fmt.Errorf("no override for this organizer")
	}
	return toOverrideResponse(o), nil
}

func (s *commissionService) DeleteOverride(ctx context.Context, organizerID uuid.UUID) error {
	return s.overrideRepo.Delete(ctx, organizerID)
}

// ResolveRateBps resolves the effective commission rate for an organizer.
// NOTE: lifetime revenue lookup will be wired in Phase 2 when Orders exist.
// For Phase 0, we use override → first active tier or default setting.
func (s *commissionService) ResolveRateBps(ctx context.Context, organizerID uuid.UUID) (int, error) {
	// 1. Explicit override wins (unless -1 which means "use tier").
	override, err := s.overrideRepo.FindByOrganizer(ctx, organizerID)
	if err != nil {
		return 0, err
	}
	if override != nil && override.RateBps >= 0 {
		return override.RateBps, nil
	}

	// 2. Tiered by lifetime revenue (placeholder revenue=0 until Phase 2).
	lifetimeRevenue := int64(0)
	tiers, err := s.tierRepo.FindAll(ctx, true)
	if err != nil {
		return 0, err
	}
	var resolved int = -1
	for _, t := range tiers {
		if lifetimeRevenue >= t.MinRevenueMinor &&
			(t.MaxRevenueMinor == nil || lifetimeRevenue < *t.MaxRevenueMinor) {
			resolved = t.RateBps
			break
		}
	}
	if resolved >= 0 {
		return resolved, nil
	}

	// 3. Fallback: default rate from settings.
	def := s.settingRepo
	_ = def // keep reference for future extension
	setting, err := s.settingRepo.FindByKey(ctx, models.SettingDefaultCommissionRate)
	if err != nil || setting == nil {
		return 500, nil // 5% safe default
	}
	var rate int
	fmt.Sscanf(setting.Value, "%d", &rate)
	return rate, nil
}

func (s *commissionService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "commission",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log:", err)
	}
}

func toCommissionTierResponse(t *models.CommissionTier) *responses.CommissionTierResponse {
	return &responses.CommissionTierResponse{
		ID:              t.ID,
		Name:            t.Name,
		MinRevenueMinor: t.MinRevenueMinor,
		MaxRevenueMinor: t.MaxRevenueMinor,
		RateBps:         t.RateBps,
		RatePercentage:  float64(t.RateBps) / 100.0,
		Priority:        t.Priority,
		IsActive:        t.IsActive,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}

func toOverrideResponse(o *models.OrganizerCommissionOverride) *responses.OrganizerCommissionOverrideResponse {
	return &responses.OrganizerCommissionOverrideResponse{
		ID:             o.ID,
		OrganizerID:    o.OrganizerID,
		RateBps:        o.RateBps,
		RatePercentage: float64(o.RateBps) / 100.0,
		Reason:         o.Reason,
		SetBy:          o.SetBy,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}
}

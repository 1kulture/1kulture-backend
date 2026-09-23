package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type partnershipConfigService struct {
	configRepo   repoInterfaces.EventPartnershipConfigRepository
	oppRepo      repoInterfaces.PartnershipOpportunityRepository
	eventRepo    repoInterfaces.EventRepository
	coOrgRepo    repoInterfaces.EventCoOrganizerRepository
	auditLogRepo repoInterfaces.AuditLogRepository
}

func NewPartnershipConfigService(
	configRepo repoInterfaces.EventPartnershipConfigRepository,
	oppRepo repoInterfaces.PartnershipOpportunityRepository,
	eventRepo repoInterfaces.EventRepository,
	coOrgRepo repoInterfaces.EventCoOrganizerRepository,
	auditLogRepo repoInterfaces.AuditLogRepository,
) serviceInterfaces.PartnershipConfigService {
	return &partnershipConfigService{
		configRepo:   configRepo,
		oppRepo:      oppRepo,
		eventRepo:    eventRepo,
		coOrgRepo:    coOrgRepo,
		auditLogRepo: auditLogRepo,
	}
}

func (s *partnershipConfigService) canManage(ctx context.Context, event *models.Event, userID uuid.UUID) bool {
	if event.OrganizerID == userID {
		return true
	}
	isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, event.ID, userID)
	return isCo
}

func (s *partnershipConfigService) UpsertConfig(ctx context.Context, actorID, eventID uuid.UUID, req *requests.EventPartnershipConfigRequest) (*responses.EventPartnershipConfigResponse, error) {
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

	cfg, err := s.configRepo.FindByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	typesJSON := toJSON(req.PartnershipTypes)
	benefitsJSON := toJSON(req.BrandBenefits)
	locationsJSON := toJSON(req.AudienceLocations)
	interestsJSON := toJSON(req.AudienceInterests)

	if cfg == nil {
		cfg = &models.EventPartnershipConfig{
			EventID:           eventID,
			IsOpen:            req.IsOpen,
			PartnershipTypes:  typesJSON,
			BrandBenefits:     benefitsJSON,
			ExpectedAttendees: req.ExpectedAttendees,
			AudienceAgeRange:  req.AudienceAgeRange,
			AudienceLocations: locationsJSON,
			AudienceInterests: interestsJSON,
			Notes:             req.Notes,
		}
		if err := s.configRepo.Create(ctx, cfg); err != nil {
			return nil, err
		}
	} else {
		cfg.IsOpen = req.IsOpen
		cfg.PartnershipTypes = typesJSON
		cfg.BrandBenefits = benefitsJSON
		cfg.ExpectedAttendees = req.ExpectedAttendees
		cfg.AudienceAgeRange = req.AudienceAgeRange
		cfg.AudienceLocations = locationsJSON
		cfg.AudienceInterests = interestsJSON
		cfg.Notes = req.Notes
		if err := s.configRepo.Update(ctx, cfg); err != nil {
			return nil, err
		}
	}

	s.audit(ctx, &actorID, "PARTNERSHIP_CONFIG_UPSERTED", eventID.String())

	// Reload with opportunities
	opps, _ := s.oppRepo.FindByEvent(ctx, eventID, false)
	return toConfigResponse(cfg, opps), nil
}

func (s *partnershipConfigService) GetConfig(ctx context.Context, eventID uuid.UUID, viewerID *uuid.UUID) (*responses.EventPartnershipConfigResponse, error) {
	cfg, err := s.configRepo.FindByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("event has no partnership config")
	}

	opps, err := s.oppRepo.FindByEvent(ctx, eventID, false)
	if err != nil {
		return nil, err
	}
	return toConfigResponse(cfg, opps), nil
}

func (s *partnershipConfigService) AddOpportunity(ctx context.Context, actorID, eventID uuid.UUID, req *requests.OpportunityCreateRequest) (*responses.OpportunityResponse, error) {
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

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	opp := &models.PartnershipOpportunity{
		EventID:        eventID,
		Title:          req.Title,
		Description:    req.Description,
		Type:           models.PartnershipType(req.Type),
		BudgetMinMinor: req.BudgetMinMinor,
		BudgetMaxMinor: req.BudgetMaxMinor,
		SlotsTotal:     req.SlotsTotal,
		SlotsRemaining: req.SlotsTotal,
		SortOrder:      req.SortOrder,
		IsActive:       isActive,
	}
	if err := s.oppRepo.Create(ctx, opp); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "OPPORTUNITY_CREATED", opp.ID.String())
	return toOpportunityResponse(opp), nil
}

func (s *partnershipConfigService) UpdateOpportunity(ctx context.Context, actorID, eventID, oppID uuid.UUID, req *requests.OpportunityUpdateRequest) (*responses.OpportunityResponse, error) {
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

	opp, err := s.oppRepo.FindByID(ctx, oppID)
	if err != nil {
		return nil, err
	}
	if opp == nil || opp.EventID != eventID {
		return nil, fmt.Errorf("opportunity not found")
	}

	if req.Title != nil {
		opp.Title = *req.Title
	}
	if req.Description != nil {
		opp.Description = *req.Description
	}
	if req.Type != nil {
		opp.Type = models.PartnershipType(*req.Type)
	}
	if req.BudgetMinMinor != nil {
		opp.BudgetMinMinor = *req.BudgetMinMinor
	}
	if req.BudgetMaxMinor != nil {
		opp.BudgetMaxMinor = *req.BudgetMaxMinor
	}
	if req.SlotsTotal != nil {
		if *req.SlotsTotal < opp.SlotsTotal-opp.SlotsRemaining {
			return nil, fmt.Errorf("cannot reduce slots below already claimed (%d)", opp.SlotsTotal-opp.SlotsRemaining)
		}
		diff := *req.SlotsTotal - opp.SlotsTotal
		opp.SlotsTotal = *req.SlotsTotal
		opp.SlotsRemaining += diff
	}
	if req.SortOrder != nil {
		opp.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		opp.IsActive = *req.IsActive
	}

	if err := s.oppRepo.Update(ctx, opp); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "OPPORTUNITY_UPDATED", opp.ID.String())
	return toOpportunityResponse(opp), nil
}

func (s *partnershipConfigService) DeleteOpportunity(ctx context.Context, actorID, eventID, oppID uuid.UUID) error {
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
	if err := s.oppRepo.Delete(ctx, oppID); err != nil {
		return err
	}
	s.audit(ctx, &actorID, "OPPORTUNITY_DELETED", oppID.String())
	return nil
}

func (s *partnershipConfigService) ListOpportunities(ctx context.Context, eventID uuid.UUID, onlyActive bool) ([]responses.OpportunityResponse, error) {
	opps, err := s.oppRepo.FindByEvent(ctx, eventID, onlyActive)
	if err != nil {
		return nil, err
	}
	out := make([]responses.OpportunityResponse, 0, len(opps))
	for i := range opps {
		out = append(out, *toOpportunityResponse(&opps[i]))
	}
	return out, nil
}

func (s *partnershipConfigService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "partnership_config",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log: ", err)
	}
}

func toConfigResponse(cfg *models.EventPartnershipConfig, opps []models.PartnershipOpportunity) *responses.EventPartnershipConfigResponse {
	r := &responses.EventPartnershipConfigResponse{
		ID:                cfg.ID,
		EventID:           cfg.EventID,
		IsOpen:            cfg.IsOpen,
		PartnershipTypes:  fromJSONStrings(cfg.PartnershipTypes),
		BrandBenefits:     fromJSONStrings(cfg.BrandBenefits),
		ExpectedAttendees: cfg.ExpectedAttendees,
		AudienceAgeRange:  cfg.AudienceAgeRange,
		AudienceLocations: fromJSONStrings(cfg.AudienceLocations),
		AudienceInterests: fromJSONStrings(cfg.AudienceInterests),
		Notes:             cfg.Notes,
		CreatedAt:         cfg.CreatedAt,
		UpdatedAt:         cfg.UpdatedAt,
	}
	if len(opps) > 0 {
		out := make([]responses.OpportunityResponse, 0, len(opps))
		for i := range opps {
			out = append(out, *toOpportunityResponse(&opps[i]))
		}
		r.Opportunities = out
	}
	return r
}

func toOpportunityResponse(o *models.PartnershipOpportunity) *responses.OpportunityResponse {
	return &responses.OpportunityResponse{
		ID:             o.ID,
		EventID:        o.EventID,
		Title:          o.Title,
		Description:    o.Description,
		Type:           string(o.Type),
		BudgetMinMinor: o.BudgetMinMinor,
		BudgetMaxMinor: o.BudgetMaxMinor,
		SlotsTotal:     o.SlotsTotal,
		SlotsRemaining: o.SlotsRemaining,
		SortOrder:      o.SortOrder,
		IsActive:       o.IsActive,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}
}

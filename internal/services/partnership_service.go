package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type partnershipService struct {
	requestRepo     repoInterfaces.PartnershipRequestRepository
	configRepo      repoInterfaces.EventPartnershipConfigRepository
	oppRepo         repoInterfaces.PartnershipOpportunityRepository
	brandRepo       repoInterfaces.BrandProfileRepository
	eventRepo       repoInterfaces.EventRepository
	coOrgRepo       repoInterfaces.EventCoOrganizerRepository
	notificationSvc serviceInterfaces.NotificationService
	auditLogRepo    repoInterfaces.AuditLogRepository
}

func NewPartnershipService(
	requestRepo repoInterfaces.PartnershipRequestRepository,
	configRepo repoInterfaces.EventPartnershipConfigRepository,
	oppRepo repoInterfaces.PartnershipOpportunityRepository,
	brandRepo repoInterfaces.BrandProfileRepository,
	eventRepo repoInterfaces.EventRepository,
	coOrgRepo repoInterfaces.EventCoOrganizerRepository,
	notificationSvc serviceInterfaces.NotificationService,
	auditLogRepo repoInterfaces.AuditLogRepository,
) serviceInterfaces.PartnershipService {
	return &partnershipService{
		requestRepo:     requestRepo,
		configRepo:      configRepo,
		oppRepo:         oppRepo,
		brandRepo:       brandRepo,
		eventRepo:       eventRepo,
		coOrgRepo:       coOrgRepo,
		notificationSvc: notificationSvc,
		auditLogRepo:    auditLogRepo,
	}
}

// ---------- Brand side ----------

func (s *partnershipService) CreateRequest(ctx context.Context, brandUserID, eventID uuid.UUID, req *requests.PartnershipRequestCreate) (*responses.PartnershipRequestResponse, error) {
	brand, err := s.brandRepo.FindByUserID(ctx, brandUserID)
	if err != nil {
		return nil, err
	}
	if brand == nil {
		return nil, fmt.Errorf("you must have a brand profile to request partnerships")
	}

	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	if event.Status != models.EventStatusPublished {
		return nil, fmt.Errorf("event is not open for partnerships")
	}

	cfg, err := s.configRepo.FindByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if cfg == nil || !cfg.IsOpen {
		return nil, fmt.Errorf("event is not currently open for partnerships")
	}

	// Prevent duplicates
	exists, err := s.requestRepo.ExistsActive(ctx, brandUserID, eventID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("you already have an active or pending request for this event")
	}

	var oppID *uuid.UUID
	if req.OpportunityID != "" {
		id, err := uuid.Parse(req.OpportunityID)
		if err != nil {
			return nil, fmt.Errorf("invalid opportunity_id")
		}
		opp, err := s.oppRepo.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if opp == nil || opp.EventID != eventID {
			return nil, fmt.Errorf("opportunity not found for this event")
		}
		if !opp.IsActive || opp.SlotsRemaining <= 0 {
			return nil, fmt.Errorf("this opportunity is no longer available")
		}
		oppID = &id
	}

	typesJSON := toJSON(req.RequestedTypes)

	pr := &models.PartnershipRequest{
		BrandProfileID: brand.ID,
		BrandUserID:    brandUserID,
		EventID:        eventID,
		OrganizerID:    event.OrganizerID,
		OpportunityID:  oppID,
		RequestedTypes: typesJSON,
		Message:        req.Message,
		Status:         models.PartnershipStatusPending,
	}
	if err := s.requestRepo.Create(ctx, pr); err != nil {
		return nil, err
	}

	// Notify organizer
	_ = s.notificationSvc.Emit(
		ctx,
		event.OrganizerID,
		models.NotificationPartnershipRequested,
		"New partnership request",
		fmt.Sprintf("%s wants to partner with %s", brand.BusinessName, event.Title),
		"partnership",
		pr.ID,
		map[string]interface{}{
			"brand_profile_id": brand.ID.String(),
			"event_id":         event.ID.String(),
		},
	)

	s.audit(ctx, &brandUserID, "PARTNERSHIP_REQUESTED", pr.ID.String())
	full, _ := s.requestRepo.FindWithRelations(ctx, pr.ID)
	return toPartnershipResponse(full), nil
}

func (s *partnershipService) ListMyBrandRequests(ctx context.Context, brandUserID uuid.UUID, status string, page, perPage int) ([]responses.PartnershipRequestResponse, int64, error) {
	filter := repoInterfaces.PartnershipRequestFilter{
		BrandUserID: &brandUserID,
		Status:      status,
		Page:        page,
		PerPage:     perPage,
	}
	list, total, err := s.requestRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.PartnershipRequestResponse, 0, len(list))
	for i := range list {
		out = append(out, *toPartnershipResponse(&list[i]))
	}
	return out, total, nil
}

func (s *partnershipService) CancelRequest(ctx context.Context, brandUserID, requestID uuid.UUID, req *requests.PartnershipDeclineRequest) (*responses.PartnershipRequestResponse, error) {
	pr, err := s.requestRepo.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, fmt.Errorf("partnership request not found")
	}
	if pr.BrandUserID != brandUserID {
		return nil, fmt.Errorf("forbidden")
	}
	if pr.Status != models.PartnershipStatusPending && pr.Status != models.PartnershipStatusAccepted {
		return nil, fmt.Errorf("request cannot be cancelled at this stage")
	}

	now := time.Now().UTC()
	pr.Status = models.PartnershipStatusCancelled
	pr.CancelledAt = &now
	pr.Reason = req.Reason
	if err := s.requestRepo.Update(ctx, pr); err != nil {
		return nil, err
	}

	// Notify organizer
	_ = s.notificationSvc.Emit(
		ctx,
		pr.OrganizerID,
		models.NotificationPartnershipCancelled,
		"Partnership request cancelled",
		"The brand cancelled their partnership request.",
		"partnership",
		pr.ID,
		nil,
	)

	s.audit(ctx, &brandUserID, "PARTNERSHIP_CANCELLED_BY_BRAND", pr.ID.String())
	full, _ := s.requestRepo.FindWithRelations(ctx, pr.ID)
	return toPartnershipResponse(full), nil
}

// ---------- Organizer side ----------

func (s *partnershipService) ListEventRequests(ctx context.Context, actorID, eventID uuid.UUID, status string, page, perPage int) ([]responses.PartnershipRequestResponse, int64, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, 0, err
	}
	if event == nil {
		return nil, 0, fmt.Errorf("event not found")
	}
	if event.OrganizerID != actorID {
		isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, eventID, actorID)
		if !isCo {
			return nil, 0, fmt.Errorf("forbidden")
		}
	}

	filter := repoInterfaces.PartnershipRequestFilter{
		EventID: &eventID,
		Status:  status,
		Page:    page,
		PerPage: perPage,
	}
	list, total, err := s.requestRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.PartnershipRequestResponse, 0, len(list))
	for i := range list {
		out = append(out, *toPartnershipResponse(&list[i]))
	}
	return out, total, nil
}

func (s *partnershipService) DecideRequest(ctx context.Context, actorID, requestID uuid.UUID, accept bool, reason string) (*responses.PartnershipRequestResponse, error) {
	pr, err := s.requestRepo.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, fmt.Errorf("partnership request not found")
	}
	if pr.Status != models.PartnershipStatusPending {
		return nil, fmt.Errorf("request has already been decided")
	}

	event, err := s.eventRepo.FindByID(ctx, pr.EventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	if event.OrganizerID != actorID {
		isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, event.ID, actorID)
		if !isCo {
			return nil, fmt.Errorf("forbidden")
		}
	}

	now := time.Now().UTC()
	if accept {
		// Reserve a slot if opportunity-specific
		if pr.OpportunityID != nil {
			if err := s.oppRepo.DecrementSlots(ctx, *pr.OpportunityID, 1); err != nil {
				return nil, err
			}
		}
		pr.Status = models.PartnershipStatusAccepted
		pr.AcceptedAt = &now
	} else {
		pr.Status = models.PartnershipStatusDeclined
		pr.DeclinedAt = &now
		pr.Reason = reason
	}
	if err := s.requestRepo.Update(ctx, pr); err != nil {
		return nil, err
	}

	notifType := models.NotificationPartnershipDeclined
	title := "Partnership declined"
	body := fmt.Sprintf("Your partnership request for %s was declined.", event.Title)
	if accept {
		notifType = models.NotificationPartnershipAccepted
		title = "Partnership accepted!"
		body = fmt.Sprintf("Your partnership request for %s has been accepted.", event.Title)
	}
	_ = s.notificationSvc.Emit(ctx, pr.BrandUserID, notifType, title, body, "partnership", pr.ID, nil)

	s.audit(ctx, &actorID, "PARTNERSHIP_DECIDED", pr.ID.String())
	full, _ := s.requestRepo.FindWithRelations(ctx, pr.ID)
	return toPartnershipResponse(full), nil
}

func (s *partnershipService) MarkActivated(ctx context.Context, actorID, requestID uuid.UUID) (*responses.PartnershipRequestResponse, error) {
	pr, err := s.requestRepo.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, fmt.Errorf("partnership not found")
	}
	if pr.OrganizerID != actorID {
		isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, pr.EventID, actorID)
		if !isCo {
			return nil, fmt.Errorf("forbidden")
		}
	}
	if pr.Status != models.PartnershipStatusAccepted {
		return nil, fmt.Errorf("partnership must be accepted before it can be activated")
	}

	now := time.Now().UTC()
	pr.Status = models.PartnershipStatusActive
	pr.ActivatedAt = &now
	if err := s.requestRepo.Update(ctx, pr); err != nil {
		return nil, err
	}

	_ = s.notificationSvc.Emit(
		ctx, pr.BrandUserID,
		models.NotificationPartnershipActivated,
		"Partnership activated",
		"The partnership is now active.",
		"partnership", pr.ID, nil,
	)

	s.audit(ctx, &actorID, "PARTNERSHIP_ACTIVATED", pr.ID.String())
	full, _ := s.requestRepo.FindWithRelations(ctx, pr.ID)
	return toPartnershipResponse(full), nil
}

func (s *partnershipService) MarkCompleted(ctx context.Context, actorID, requestID uuid.UUID) (*responses.PartnershipRequestResponse, error) {
	pr, err := s.requestRepo.FindByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, fmt.Errorf("partnership not found")
	}
	if pr.OrganizerID != actorID {
		isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, pr.EventID, actorID)
		if !isCo {
			return nil, fmt.Errorf("forbidden")
		}
	}
	if pr.Status != models.PartnershipStatusActive && pr.Status != models.PartnershipStatusAccepted {
		return nil, fmt.Errorf("only accepted or active partnerships can be completed")
	}

	now := time.Now().UTC()
	pr.Status = models.PartnershipStatusCompleted
	pr.CompletedAt = &now
	if err := s.requestRepo.Update(ctx, pr); err != nil {
		return nil, err
	}

	_ = s.notificationSvc.Emit(
		ctx, pr.BrandUserID,
		models.NotificationPartnershipCompleted,
		"Partnership completed",
		"Thanks for partnering with us!",
		"partnership", pr.ID, nil,
	)

	s.audit(ctx, &actorID, "PARTNERSHIP_COMPLETED", pr.ID.String())
	full, _ := s.requestRepo.FindWithRelations(ctx, pr.ID)
	return toPartnershipResponse(full), nil
}

// ---------- Shared ----------

func (s *partnershipService) GetRequest(ctx context.Context, viewerID, requestID uuid.UUID) (*responses.PartnershipRequestResponse, error) {
	pr, err := s.requestRepo.FindWithRelations(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, fmt.Errorf("partnership not found")
	}
	if pr.BrandUserID != viewerID && pr.OrganizerID != viewerID {
		isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, pr.EventID, viewerID)
		if !isCo {
			return nil, fmt.Errorf("forbidden")
		}
	}
	return toPartnershipResponse(pr), nil
}

func (s *partnershipService) ListAll(ctx context.Context, filter repoInterfaces.PartnershipRequestFilter, viewerID uuid.UUID) ([]responses.PartnershipRequestResponse, int64, error) {
	// This is used by admin; enforce caller's role upstream.
	list, total, err := s.requestRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.PartnershipRequestResponse, 0, len(list))
	for i := range list {
		out = append(out, *toPartnershipResponse(&list[i]))
	}
	return out, total, nil
}

func (s *partnershipService) Dashboard(ctx context.Context, viewerID uuid.UUID, eventID *uuid.UUID) (*responses.PartnershipDashboardResponse, error) {
	filter := repoInterfaces.PartnershipRequestFilter{}
	if eventID != nil {
		// Ensure the viewer can see this event's data
		event, err := s.eventRepo.FindByID(ctx, *eventID)
		if err != nil {
			return nil, err
		}
		if event == nil {
			return nil, fmt.Errorf("event not found")
		}
		if event.OrganizerID != viewerID {
			isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, *eventID, viewerID)
			if !isCo {
				return nil, fmt.Errorf("forbidden")
			}
		}
		filter.EventID = eventID
	} else {
		filter.OrganizerID = &viewerID
	}

	counts, err := s.requestRepo.CountByStatus(ctx, filter)
	if err != nil {
		return nil, err
	}
	d := &responses.PartnershipDashboardResponse{
		PendingCount:   counts[string(models.PartnershipStatusPending)],
		AcceptedCount:  counts[string(models.PartnershipStatusAccepted)],
		ActiveCount:    counts[string(models.PartnershipStatusActive)],
		CompletedCount: counts[string(models.PartnershipStatusCompleted)],
		DeclinedCount:  counts[string(models.PartnershipStatusDeclined)],
		CancelledCount: counts[string(models.PartnershipStatusCancelled)],
	}
	for _, c := range counts {
		d.TotalCount += c
	}
	return d, nil
}

// ---------- helpers ----------

func (s *partnershipService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "partnership",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log: ", err)
	}
}

func toPartnershipResponse(pr *models.PartnershipRequest) *responses.PartnershipRequestResponse {
	if pr == nil {
		return nil
	}
	r := &responses.PartnershipRequestResponse{
		ID:             pr.ID,
		BrandProfileID: pr.BrandProfileID,
		BrandUserID:    pr.BrandUserID,
		EventID:        pr.EventID,
		OrganizerID:    pr.OrganizerID,
		OpportunityID:  pr.OpportunityID,
		RequestedTypes: fromJSONStrings(pr.RequestedTypes),
		Message:        pr.Message,
		Status:         string(pr.Status),
		AcceptedAt:     pr.AcceptedAt,
		ActivatedAt:    pr.ActivatedAt,
		CompletedAt:    pr.CompletedAt,
		DeclinedAt:     pr.DeclinedAt,
		CancelledAt:    pr.CancelledAt,
		Reason:         pr.Reason,
		CreatedAt:      pr.CreatedAt,
		UpdatedAt:      pr.UpdatedAt,
	}

	// Embed brand summary if loaded
	if pr.BrandProfile.ID != uuid.Nil {
		r.Brand = toBrandSummary(&pr.BrandProfile)
	}

	// Embed event reference if loaded
	if pr.Event.ID != uuid.Nil {
		r.Event = &responses.PartnershipEventRef{
			ID:           pr.Event.ID,
			Slug:         pr.Event.Slug,
			Title:        pr.Event.Title,
			ThumbnailURL: pr.Event.ThumbnailURL,
			StartAt:      pr.Event.StartAt,
			EndAt:        pr.Event.EndAt,
			Currency:     pr.Event.Currency,
		}
	}
	return r
}

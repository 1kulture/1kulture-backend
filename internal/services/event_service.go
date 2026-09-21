package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type eventService struct {
	eventRepo       interfaces.EventRepository
	occRepo         interfaces.EventOccurrenceRepository
	coOrgRepo       interfaces.EventCoOrganizerRepository
	staffRepo       interfaces.EventStaffRepository
	followerRepo    interfaces.EventFollowerRepository
	orgFollowerRepo interfaces.OrganizerFollowerRepository
	shareRepo       interfaces.EventShareRepository
	currencyRepo    interfaces.CurrencyRepository
	settingService  serviceInterfaces.SettingService
	auditLogRepo    interfaces.AuditLogRepository
}

func NewEventService(
	eventRepo interfaces.EventRepository,
	occRepo interfaces.EventOccurrenceRepository,
	coOrgRepo interfaces.EventCoOrganizerRepository,
	staffRepo interfaces.EventStaffRepository,
	followerRepo interfaces.EventFollowerRepository,
	orgFollowerRepo interfaces.OrganizerFollowerRepository,
	shareRepo interfaces.EventShareRepository,
	currencyRepo interfaces.CurrencyRepository,
	settingService serviceInterfaces.SettingService,
	auditLogRepo interfaces.AuditLogRepository,
) serviceInterfaces.EventService {
	return &eventService{
		eventRepo:       eventRepo,
		occRepo:         occRepo,
		coOrgRepo:       coOrgRepo,
		staffRepo:       staffRepo,
		followerRepo:    followerRepo,
		orgFollowerRepo: orgFollowerRepo,
		shareRepo:       shareRepo,
		currencyRepo:    currencyRepo,
		settingService:  settingService,
		auditLogRepo:    auditLogRepo,
	}
}

// ---------- CRUD ----------

func (s *eventService) Create(ctx context.Context, actorID uuid.UUID, req *requests.EventCreateRequest) (*responses.EventResponse, error) {
	// Validate currency is enabled
	currency := req.Currency
	if currency == "" {
		def, err := s.currencyRepo.FindDefault(ctx)
		if err != nil {
			return nil, err
		}
		if def == nil {
			return nil, fmt.Errorf("no default currency configured")
		}
		currency = def.Code
	} else {
		c, err := s.currencyRepo.FindByCode(ctx, currency)
		if err != nil {
			return nil, err
		}
		if c == nil || !c.IsEnabled {
			return nil, fmt.Errorf("currency '%s' is not enabled", currency)
		}
	}

	// Validate times
	if !req.EndAt.After(req.StartAt) {
		return nil, fmt.Errorf("end_at must be after start_at")
	}
	if req.EventType == "in_person" || req.EventType == "hybrid" {
		if req.VenueName == "" || req.VenueCity == "" {
			return nil, fmt.Errorf("venue_name and venue_city are required for in_person/hybrid events")
		}
	}
	if req.EventType == "virtual" || req.EventType == "hybrid" {
		if req.VirtualURL == "" {
			return nil, fmt.Errorf("virtual_url is required for virtual/hybrid events")
		}
	}

	slug, err := s.generateUniqueSlug(ctx, req.Title)
	if err != nil {
		return nil, err
	}

	event := &models.Event{
		OrganizerID:      actorID,
		Slug:             slug,
		Title:            req.Title,
		Summary:          req.Summary,
		Description:      req.Description,
		BannerURL:        req.BannerURL,
		ThumbnailURL:     req.ThumbnailURL,
		Gallery:          toJSON(req.Gallery),
		Category:         req.Category,
		Tags:             toJSON(req.Tags),
		EventType:        models.EventType(req.EventType),
		VenueName:        req.VenueName,
		VenueAddress:     req.VenueAddress,
		VenueCity:        req.VenueCity,
		VenueState:       req.VenueState,
		VenueCountry:     req.VenueCountry,
		VenuePostalCode:  req.VenuePostalCode,
		VenueLat:         req.VenueLat,
		VenueLng:         req.VenueLng,
		VirtualURL:       req.VirtualURL,
		VirtualPlatform:  req.VirtualPlatform,
		Timezone:         req.Timezone,
		StartAt:          req.StartAt,
		EndAt:            req.EndAt,
		DoorsOpenAt:      req.DoorsOpenAt,
		Capacity:         req.Capacity,
		AgeRestriction:   models.AgeRestriction(defaultStr(req.AgeRestriction, string(models.AgeAllAges))),
		Status:           models.EventStatusDraft,
		Visibility:       models.EventVisibility(defaultStr(req.Visibility, string(models.EventVisibilityPublic))),
		RefundPolicy:     req.RefundPolicy,
		RefundPolicyDays: defaultInt(req.RefundPolicyDays, s.settingService.GetInt(ctx, models.SettingRefundPolicyDefaultDays, 7)),
		Currency:         currency,
	}

	if event.RefundPolicy == "" {
		event.RefundPolicy = s.settingService.GetString(ctx, models.SettingRefundPolicyDefaultText, "")
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, err
	}

	s.audit(ctx, &actorID, "EVENT_CREATED", event.ID.String())
	return s.buildResponse(ctx, event, &actorID), nil
}

func (s *eventService) GetByID(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*responses.EventResponse, error) {
	event, err := s.eventRepo.FindWithRelations(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	// If unlisted/private/draft, only organizer/coorg can view
	if !s.canView(ctx, event, viewerID) {
		return nil, fmt.Errorf("event not found")
	}
	return s.buildResponse(ctx, event, viewerID), nil
}

func (s *eventService) GetBySlug(ctx context.Context, slug string, viewerID *uuid.UUID) (*responses.EventResponse, error) {
	event, err := s.eventRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	// reload with relations
	full, err := s.eventRepo.FindWithRelations(ctx, event.ID)
	if err != nil {
		return nil, err
	}
	if !s.canView(ctx, full, viewerID) {
		return nil, fmt.Errorf("event not found")
	}
	return s.buildResponse(ctx, full, viewerID), nil
}

func (s *eventService) Update(ctx context.Context, actorID, eventID uuid.UUID, req *requests.EventUpdateRequest) (*responses.EventResponse, error) {
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

	if req.Title != nil {
		event.Title = *req.Title
	}
	if req.Summary != nil {
		event.Summary = *req.Summary
	}
	if req.Description != nil {
		event.Description = *req.Description
	}
	if req.BannerURL != nil {
		event.BannerURL = *req.BannerURL
	}
	if req.ThumbnailURL != nil {
		event.ThumbnailURL = *req.ThumbnailURL
	}
	if req.Gallery != nil {
		event.Gallery = toJSON(req.Gallery)
	}
	if req.Category != nil {
		event.Category = *req.Category
	}
	if req.Tags != nil {
		event.Tags = toJSON(req.Tags)
	}
	if req.EventType != nil {
		event.EventType = models.EventType(*req.EventType)
	}
	if req.VenueName != nil {
		event.VenueName = *req.VenueName
	}
	if req.VenueAddress != nil {
		event.VenueAddress = *req.VenueAddress
	}
	if req.VenueCity != nil {
		event.VenueCity = *req.VenueCity
	}
	if req.VenueState != nil {
		event.VenueState = *req.VenueState
	}
	if req.VenueCountry != nil {
		event.VenueCountry = *req.VenueCountry
	}
	if req.VenuePostalCode != nil {
		event.VenuePostalCode = *req.VenuePostalCode
	}
	if req.VenueLat != nil {
		event.VenueLat = req.VenueLat
	}
	if req.VenueLng != nil {
		event.VenueLng = req.VenueLng
	}
	if req.VirtualURL != nil {
		event.VirtualURL = *req.VirtualURL
	}
	if req.VirtualPlatform != nil {
		event.VirtualPlatform = *req.VirtualPlatform
	}
	if req.Timezone != nil {
		event.Timezone = *req.Timezone
	}
	if req.StartAt != nil {
		event.StartAt = *req.StartAt
	}
	if req.EndAt != nil {
		event.EndAt = *req.EndAt
	}
	if req.DoorsOpenAt != nil {
		event.DoorsOpenAt = req.DoorsOpenAt
	}
	if req.Capacity != nil {
		event.Capacity = req.Capacity
	}
	if req.AgeRestriction != nil {
		event.AgeRestriction = models.AgeRestriction(*req.AgeRestriction)
	}
	if req.Visibility != nil {
		event.Visibility = models.EventVisibility(*req.Visibility)
	}
	if req.RefundPolicy != nil {
		event.RefundPolicy = *req.RefundPolicy
	}
	if req.RefundPolicyDays != nil {
		event.RefundPolicyDays = *req.RefundPolicyDays
	}
	if req.Currency != nil {
		c, err := s.currencyRepo.FindByCode(ctx, *req.Currency)
		if err != nil {
			return nil, err
		}
		if c == nil || !c.IsEnabled {
			return nil, fmt.Errorf("currency '%s' is not enabled", *req.Currency)
		}
		event.Currency = *req.Currency
	}

	if !event.EndAt.After(event.StartAt) {
		return nil, fmt.Errorf("end_at must be after start_at")
	}

	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "EVENT_UPDATED", event.ID.String())
	full, _ := s.eventRepo.FindWithRelations(ctx, event.ID)
	return s.buildResponse(ctx, full, &actorID), nil
}

func (s *eventService) Delete(ctx context.Context, actorID, eventID uuid.UUID) error {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return err
	}
	if event == nil {
		return fmt.Errorf("event not found")
	}
	if event.OrganizerID != actorID {
		return fmt.Errorf("forbidden")
	}
	if err := s.eventRepo.Delete(ctx, eventID); err != nil {
		return err
	}
	s.audit(ctx, &actorID, "EVENT_DELETED", eventID.String())
	return nil
}

func (s *eventService) List(ctx context.Context, filter interfaces.EventListFilter, viewerID *uuid.UUID) ([]responses.EventSummaryResponse, int64, error) {
	// If filtering by organizer and viewer is that organizer, include drafts.
	// Otherwise force published.
	if filter.OrganizerID == nil || viewerID == nil || *filter.OrganizerID != *viewerID {
		if filter.Status == "" || filter.Status == string(models.EventStatusDraft) {
			filter.Status = string(models.EventStatusPublished)
		}
	}

	events, total, err := s.eventRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.EventSummaryResponse, 0, len(events))
	for i := range events {
		out = append(out, *toEventSummary(&events[i]))
	}
	return out, total, nil
}

// ---------- Lifecycle ----------

func (s *eventService) Publish(ctx context.Context, actorID, eventID uuid.UUID) (*responses.EventResponse, error) {
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
	if event.Status == models.EventStatusCancelled {
		return nil, fmt.Errorf("cannot publish a cancelled event")
	}
	now := time.Now()
	event.Status = models.EventStatusPublished
	event.PublishedAt = &now
	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "EVENT_PUBLISHED", eventID.String())
	full, _ := s.eventRepo.FindWithRelations(ctx, eventID)
	return s.buildResponse(ctx, full, &actorID), nil
}

func (s *eventService) Unpublish(ctx context.Context, actorID, eventID uuid.UUID) (*responses.EventResponse, error) {
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
	event.Status = models.EventStatusDraft
	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "EVENT_UNPUBLISHED", eventID.String())
	full, _ := s.eventRepo.FindWithRelations(ctx, eventID)
	return s.buildResponse(ctx, full, &actorID), nil
}

func (s *eventService) Cancel(ctx context.Context, actorID, eventID uuid.UUID, req *requests.EventCancelRequest) (*responses.EventResponse, error) {
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
	now := time.Now()
	event.Status = models.EventStatusCancelled
	event.CancelledAt = &now
	event.CancelReason = req.Reason
	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "EVENT_CANCELLED", eventID.String())
	full, _ := s.eventRepo.FindWithRelations(ctx, eventID)
	return s.buildResponse(ctx, full, &actorID), nil
}

func (s *eventService) Postpone(ctx context.Context, actorID, eventID uuid.UUID, req *requests.EventPostponeRequest) (*responses.EventResponse, error) {
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
	if !req.NewEndAt.After(req.NewStartAt) {
		return nil, fmt.Errorf("new_end_at must be after new_start_at")
	}
	event.Status = models.EventStatusPostponed
	event.StartAt = req.NewStartAt
	event.EndAt = req.NewEndAt
	event.CancelReason = req.Reason
	if err := s.eventRepo.Update(ctx, event); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "EVENT_POSTPONED", eventID.String())
	full, _ := s.eventRepo.FindWithRelations(ctx, eventID)
	return s.buildResponse(ctx, full, &actorID), nil
}

// ---------- Occurrences ----------

func (s *eventService) AddOccurrence(ctx context.Context, actorID, eventID uuid.UUID, req *requests.OccurrenceCreateRequest) (*responses.EventOccurrenceResponse, error) {
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
	if !req.EndAt.After(req.StartAt) {
		return nil, fmt.Errorf("end_at must be after start_at")
	}
	occ := &models.EventOccurrence{
		EventID:      eventID,
		SessionName:  req.SessionName,
		StartAt:      req.StartAt,
		EndAt:        req.EndAt,
		VenueName:    req.VenueName,
		VenueAddress: req.VenueAddress,
		VirtualURL:   req.VirtualURL,
		Capacity:     req.Capacity,
		SortOrder:    req.SortOrder,
	}
	if err := s.occRepo.Create(ctx, occ); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "EVENT_OCCURRENCE_ADDED", occ.ID.String())
	return toOccurrenceResponse(occ), nil
}

func (s *eventService) UpdateOccurrence(ctx context.Context, actorID, eventID, occID uuid.UUID, req *requests.OccurrenceUpdateRequest) (*responses.EventOccurrenceResponse, error) {
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
	occ, err := s.occRepo.FindByID(ctx, occID)
	if err != nil {
		return nil, err
	}
	if occ == nil || occ.EventID != eventID {
		return nil, fmt.Errorf("occurrence not found")
	}
	if req.SessionName != nil {
		occ.SessionName = *req.SessionName
	}
	if req.StartAt != nil {
		occ.StartAt = *req.StartAt
	}
	if req.EndAt != nil {
		occ.EndAt = *req.EndAt
	}
	if req.VenueName != nil {
		occ.VenueName = *req.VenueName
	}
	if req.VenueAddress != nil {
		occ.VenueAddress = *req.VenueAddress
	}
	if req.VirtualURL != nil {
		occ.VirtualURL = *req.VirtualURL
	}
	if req.Capacity != nil {
		occ.Capacity = req.Capacity
	}
	if req.SortOrder != nil {
		occ.SortOrder = *req.SortOrder
	}
	if !occ.EndAt.After(occ.StartAt) {
		return nil, fmt.Errorf("end_at must be after start_at")
	}
	if err := s.occRepo.Update(ctx, occ); err != nil {
		return nil, err
	}
	return toOccurrenceResponse(occ), nil
}

func (s *eventService) DeleteOccurrence(ctx context.Context, actorID, eventID, occID uuid.UUID) error {
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
	return s.occRepo.Delete(ctx, occID)
}

func (s *eventService) ListOccurrences(ctx context.Context, eventID uuid.UUID) ([]responses.EventOccurrenceResponse, error) {
	list, err := s.occRepo.FindByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	out := make([]responses.EventOccurrenceResponse, 0, len(list))
	for i := range list {
		out = append(out, *toOccurrenceResponse(&list[i]))
	}
	return out, nil
}

// ---------- Team: Co-Organizers ----------

func (s *eventService) AddCoOrganizer(ctx context.Context, actorID, eventID uuid.UUID, req *requests.CoOrganizerAddRequest) (*responses.EventCoOrganizerResponse, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	if event.OrganizerID != actorID {
		return nil, fmt.Errorf("only the event owner can add co-organizers")
	}
	targetID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}
	existing, err := s.coOrgRepo.FindByEventAndUser(ctx, eventID, targetID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("user is already a co-organizer")
	}
	co := &models.EventCoOrganizer{
		EventID:     eventID,
		UserID:      targetID,
		Role:        models.EventCoOrganizerRole(req.Role),
		Permissions: toJSON(req.Permissions),
	}
	if err := s.coOrgRepo.Create(ctx, co); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "CO_ORGANIZER_ADDED", targetID.String())
	fresh, _ := s.coOrgRepo.FindByEventAndUser(ctx, eventID, targetID)
	return toCoOrgResponse(fresh), nil
}

func (s *eventService) RemoveCoOrganizer(ctx context.Context, actorID, eventID, targetUserID uuid.UUID) error {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return err
	}
	if event == nil {
		return fmt.Errorf("event not found")
	}
	if event.OrganizerID != actorID {
		return fmt.Errorf("only the event owner can remove co-organizers")
	}
	if err := s.coOrgRepo.Delete(ctx, eventID, targetUserID); err != nil {
		return err
	}
	s.audit(ctx, &actorID, "CO_ORGANIZER_REMOVED", targetUserID.String())
	return nil
}

func (s *eventService) ListCoOrganizers(ctx context.Context, eventID uuid.UUID) ([]responses.EventCoOrganizerResponse, error) {
	list, err := s.coOrgRepo.FindByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	out := make([]responses.EventCoOrganizerResponse, 0, len(list))
	for i := range list {
		out = append(out, *toCoOrgResponse(&list[i]))
	}
	return out, nil
}

// ---------- Team: Staff ----------

func (s *eventService) AddStaff(ctx context.Context, actorID, eventID uuid.UUID, req *requests.StaffAddRequest) (*responses.EventStaffResponse, error) {
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
	targetID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}
	existing, err := s.staffRepo.FindByEventAndUser(ctx, eventID, targetID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("user is already staff")
	}
	st := &models.EventStaff{
		EventID: eventID,
		UserID:  targetID,
		Role:    models.EventStaffRole(req.Role),
	}
	if err := s.staffRepo.Create(ctx, st); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "STAFF_ADDED", targetID.String())
	fresh, _ := s.staffRepo.FindByEventAndUser(ctx, eventID, targetID)
	return toStaffResponse(fresh), nil
}

func (s *eventService) RemoveStaff(ctx context.Context, actorID, eventID, targetUserID uuid.UUID) error {
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
	if err := s.staffRepo.Delete(ctx, eventID, targetUserID); err != nil {
		return err
	}
	s.audit(ctx, &actorID, "STAFF_REMOVED", targetUserID.String())
	return nil
}

func (s *eventService) ListStaff(ctx context.Context, eventID uuid.UUID) ([]responses.EventStaffResponse, error) {
	list, err := s.staffRepo.FindByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	out := make([]responses.EventStaffResponse, 0, len(list))
	for i := range list {
		out = append(out, *toStaffResponse(&list[i]))
	}
	return out, nil
}

// ---------- Follow ----------

func (s *eventService) FollowEvent(ctx context.Context, userID, eventID uuid.UUID) error {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return err
	}
	if event == nil {
		return fmt.Errorf("event not found")
	}
	already, err := s.followerRepo.IsFollowing(ctx, eventID, userID)
	if err != nil {
		return err
	}
	if already {
		return nil
	}
	f := &models.EventFollower{
		EventID:           eventID,
		UserID:            userID,
		NotifyNewSessions: true,
		NotifyUpdates:     true,
	}
	if err := s.followerRepo.Follow(ctx, f); err != nil {
		return err
	}
	s.audit(ctx, &userID, "EVENT_FOLLOWED", eventID.String())
	return nil
}

func (s *eventService) UnfollowEvent(ctx context.Context, userID, eventID uuid.UUID) error {
	if err := s.followerRepo.Unfollow(ctx, eventID, userID); err != nil {
		return err
	}
	s.audit(ctx, &userID, "EVENT_UNFOLLOWED", eventID.String())
	return nil
}

func (s *eventService) ListEventFollowers(ctx context.Context, actorID, eventID uuid.UUID, page, perPage int) ([]responses.EventFollowerResponse, int64, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, 0, err
	}
	if event == nil {
		return nil, 0, fmt.Errorf("event not found")
	}
	if !s.canManage(ctx, event, actorID) {
		return nil, 0, fmt.Errorf("forbidden")
	}
	list, total, err := s.followerRepo.FindByEvent(ctx, eventID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.EventFollowerResponse, 0, len(list))
	for i := range list {
		out = append(out, responses.EventFollowerResponse{
			UserID:            list[i].UserID,
			FirstName:         list[i].User.FirstName,
			LastName:          list[i].User.LastName,
			AvatarURL:         list[i].User.AvatarURL,
			NotifyNewSessions: list[i].NotifyNewSessions,
			NotifyUpdates:     list[i].NotifyUpdates,
			FollowedAt:        list[i].CreatedAt,
		})
	}
	return out, total, nil
}

func (s *eventService) FollowOrganizer(ctx context.Context, userID, organizerID uuid.UUID) error {
	already, err := s.orgFollowerRepo.IsFollowing(ctx, organizerID, userID)
	if err != nil {
		return err
	}
	if already {
		return nil
	}
	f := &models.EventOrganizerFollower{
		UserID:          userID,
		OrganizerID:     organizerID,
		NotifyNewEvents: true,
	}
	if err := s.orgFollowerRepo.Follow(ctx, f); err != nil {
		return err
	}
	s.audit(ctx, &userID, "ORGANIZER_FOLLOWED", organizerID.String())
	return nil
}

func (s *eventService) UnfollowOrganizer(ctx context.Context, userID, organizerID uuid.UUID) error {
	if err := s.orgFollowerRepo.Unfollow(ctx, organizerID, userID); err != nil {
		return err
	}
	s.audit(ctx, &userID, "ORGANIZER_UNFOLLOWED", organizerID.String())
	return nil
}

// ---------- Share ----------

func (s *eventService) RecordShare(ctx context.Context, eventID uuid.UUID, userID *uuid.UUID, req *requests.ShareRequest, ip, userAgent string) (*responses.EventShareResponse, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	ref := ""
	if userID != nil {
		ref = userID.String()[:8]
	}
	share := &models.EventShare{
		EventID:      eventID,
		UserID:       userID,
		Channel:      models.ShareChannel(req.Channel),
		ReferralCode: ref,
		IPAddress:    ip,
		UserAgent:    userAgent,
	}
	if err := s.shareRepo.Create(ctx, share); err != nil {
		return nil, err
	}
	return &responses.EventShareResponse{
		ID:           share.ID,
		EventID:      share.EventID,
		Channel:      string(share.Channel),
		ReferralCode: share.ReferralCode,
		SharedAt:     share.CreatedAt,
	}, nil
}

// ---------- Helpers ----------

func (s *eventService) canManage(ctx context.Context, event *models.Event, userID uuid.UUID) bool {
	if event.OrganizerID == userID {
		return true
	}
	isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, event.ID, userID)
	return isCo
}

func (s *eventService) canView(ctx context.Context, event *models.Event, viewerID *uuid.UUID) bool {
	if event.Status == models.EventStatusPublished && event.Visibility == models.EventVisibilityPublic {
		return true
	}
	if viewerID == nil {
		return false
	}
	return s.canManage(ctx, event, *viewerID)
}

func (s *eventService) buildResponse(ctx context.Context, e *models.Event, viewerID *uuid.UUID) *responses.EventResponse {
	if e == nil {
		return nil
	}
	followerCount, _ := s.followerRepo.CountByEvent(ctx, e.ID)
	shareCount, _ := s.shareRepo.CountByEvent(ctx, e.ID)

	isFollowing := false
	if viewerID != nil {
		isFollowing, _ = s.followerRepo.IsFollowing(ctx, e.ID, *viewerID)
	}

	res := &responses.EventResponse{
		ID:               e.ID,
		OrganizerID:      e.OrganizerID,
		Slug:             e.Slug,
		Title:            e.Title,
		Summary:          e.Summary,
		Description:      e.Description,
		BannerURL:        e.BannerURL,
		ThumbnailURL:     e.ThumbnailURL,
		Gallery:          fromJSONStrings(e.Gallery),
		Category:         e.Category,
		Tags:             fromJSONStrings(e.Tags),
		EventType:        string(e.EventType),
		VenueName:        e.VenueName,
		VenueAddress:     e.VenueAddress,
		VenueCity:        e.VenueCity,
		VenueState:       e.VenueState,
		VenueCountry:     e.VenueCountry,
		VenuePostalCode:  e.VenuePostalCode,
		VenueLat:         e.VenueLat,
		VenueLng:         e.VenueLng,
		VirtualURL:       e.VirtualURL,
		VirtualPlatform:  e.VirtualPlatform,
		Timezone:         e.Timezone,
		StartAt:          e.StartAt,
		EndAt:            e.EndAt,
		DoorsOpenAt:      e.DoorsOpenAt,
		Capacity:         e.Capacity,
		AgeRestriction:   string(e.AgeRestriction),
		Status:           string(e.Status),
		Visibility:       string(e.Visibility),
		IsFeatured:       e.IsFeatured,
		RefundPolicy:     e.RefundPolicy,
		RefundPolicyDays: e.RefundPolicyDays,
		Currency:         e.Currency,
		PublishedAt:      e.PublishedAt,
		CancelledAt:      e.CancelledAt,
		CancelReason:     e.CancelReason,
		FollowerCount:    int(followerCount),
		ShareCount:       int(shareCount),
		IsFollowing:      isFollowing,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}

	if len(e.Occurrences) > 0 {
		occs := make([]responses.EventOccurrenceResponse, 0, len(e.Occurrences))
		for i := range e.Occurrences {
			occs = append(occs, *toOccurrenceResponse(&e.Occurrences[i]))
		}
		res.Occurrences = occs
	}

	if len(e.CoOrganizers) > 0 {
		cos := make([]responses.EventCoOrganizerResponse, 0, len(e.CoOrganizers))
		for i := range e.CoOrganizers {
			cos = append(cos, *toCoOrgResponse(&e.CoOrganizers[i]))
		}
		res.CoOrganizers = cos
	}

	if e.Organizer.ID != uuid.Nil {
		res.Organizer = &responses.EventOrganizerBrief{
			ID:        e.Organizer.ID,
			FirstName: e.Organizer.FirstName,
			LastName:  e.Organizer.LastName,
			AvatarURL: e.Organizer.AvatarURL,
		}
	}

	return res
}

func (s *eventService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "event",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log:", err)
	}
}

func (s *eventService) generateUniqueSlug(ctx context.Context, title string) (string, error) {
	base := slugify(title)
	if base == "" {
		base = "event"
	}
	candidate := base
	for i := 1; i < 20; i++ {
		exists, err := s.eventRepo.SlugExists(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
	// fallback with random suffix
	return fmt.Sprintf("%s-%s", base, uuid.New().String()[:8]), nil
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		case r == ' ' || r == '-' || r == '_':
			if !prevDash {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	out := b.String()
	return strings.Trim(out, "-")
}

func toJSON(v interface{}) datatypes.JSON {
	if v == nil {
		return datatypes.JSON([]byte("null"))
	}
	return datatypes.JSON(mustMarshal(v))
}

func fromJSONStrings(j datatypes.JSON) []string {
	if len(j) == 0 || string(j) == "null" {
		return nil
	}
	var out []string
	_ = jsonUnmarshal(j, &out)
	return out
}

func defaultStr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func defaultInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}

func toEventSummary(e *models.Event) *responses.EventSummaryResponse {
	return &responses.EventSummaryResponse{
		ID:           e.ID,
		Slug:         e.Slug,
		Title:        e.Title,
		Summary:      e.Summary,
		ThumbnailURL: e.ThumbnailURL,
		BannerURL:    e.BannerURL,
		Category:     e.Category,
		EventType:    string(e.EventType),
		VenueCity:    e.VenueCity,
		VenueCountry: e.VenueCountry,
		Timezone:     e.Timezone,
		StartAt:      e.StartAt,
		EndAt:        e.EndAt,
		Status:       string(e.Status),
		Visibility:   string(e.Visibility),
		Currency:     e.Currency,
		IsFeatured:   e.IsFeatured,
		OrganizerID:  e.OrganizerID,
		CreatedAt:    e.CreatedAt,
	}
}

func toOccurrenceResponse(o *models.EventOccurrence) *responses.EventOccurrenceResponse {
	return &responses.EventOccurrenceResponse{
		ID:           o.ID,
		EventID:      o.EventID,
		SessionName:  o.SessionName,
		StartAt:      o.StartAt,
		EndAt:        o.EndAt,
		VenueName:    o.VenueName,
		VenueAddress: o.VenueAddress,
		VirtualURL:   o.VirtualURL,
		Capacity:     o.Capacity,
		SortOrder:    o.SortOrder,
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
	}
}

func toCoOrgResponse(c *models.EventCoOrganizer) *responses.EventCoOrganizerResponse {
	if c == nil {
		return nil
	}
	r := &responses.EventCoOrganizerResponse{
		ID:          c.ID,
		EventID:     c.EventID,
		UserID:      c.UserID,
		Role:        string(c.Role),
		Permissions: fromJSONMap(c.Permissions),
		CreatedAt:   c.CreatedAt,
	}
	if c.User.ID != uuid.Nil {
		r.FirstName = c.User.FirstName
		r.LastName = c.User.LastName
		r.Email = c.User.Email
		r.AvatarURL = c.User.AvatarURL
	}
	return r
}

func toStaffResponse(st *models.EventStaff) *responses.EventStaffResponse {
	if st == nil {
		return nil
	}
	r := &responses.EventStaffResponse{
		ID:        st.ID,
		EventID:   st.EventID,
		UserID:    st.UserID,
		Role:      string(st.Role),
		CreatedAt: st.CreatedAt,
	}
	if st.User.ID != uuid.Nil {
		r.FirstName = st.User.FirstName
		r.LastName = st.User.LastName
		r.Email = st.User.Email
	}
	return r
}

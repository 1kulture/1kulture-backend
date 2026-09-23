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

type ticketTypeService struct {
	ttRepo       repoInterfaces.TicketTypeRepository
	eventRepo    repoInterfaces.EventRepository
	coOrgRepo    repoInterfaces.EventCoOrganizerRepository
	auditLogRepo repoInterfaces.AuditLogRepository
}

func NewTicketTypeService(
	ttRepo repoInterfaces.TicketTypeRepository,
	eventRepo repoInterfaces.EventRepository,
	coOrgRepo repoInterfaces.EventCoOrganizerRepository,
	auditLogRepo repoInterfaces.AuditLogRepository,
) serviceInterfaces.TicketTypeService {
	return &ticketTypeService{
		ttRepo:       ttRepo,
		eventRepo:    eventRepo,
		coOrgRepo:    coOrgRepo,
		auditLogRepo: auditLogRepo,
	}
}

func (s *ticketTypeService) canManage(ctx context.Context, event *models.Event, userID uuid.UUID) bool {
	if event.OrganizerID == userID {
		return true
	}
	isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, event.ID, userID)
	return isCo
}

func (s *ticketTypeService) Create(ctx context.Context, actorID, eventID uuid.UUID, req *requests.TicketTypeCreateRequest) (*responses.TicketTypeResponse, error) {
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
	if event.Status == models.EventStatusCancelled || event.Status == models.EventStatusCompleted {
		return nil, fmt.Errorf("cannot add ticket types to a %s event", event.Status)
	}

	var occID *uuid.UUID
	if req.OccurrenceID != nil && *req.OccurrenceID != "" {
		id, err := uuid.Parse(*req.OccurrenceID)
		if err != nil {
			return nil, fmt.Errorf("invalid occurrence_id")
		}
		occID = &id
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	accessLevel := req.AccessLevel
	if accessLevel == "" {
		accessLevel = "general"
	}

	tt := &models.TicketType{
		EventID:          eventID,
		OccurrenceID:     occID,
		Name:             req.Name,
		Description:      req.Description,
		ImageURL:         req.ImageURL,
		PriceMinor:       req.PriceMinor,
		Currency:         event.Currency,
		QuantityTotal:    req.QuantityTotal,
		PerUserLimit:     req.PerUserLimit,
		SalesStartAt:     req.SalesStartAt,
		SalesEndAt:       req.SalesEndAt,
		IsHidden:         req.IsHidden,
		IsActive:         isActive,
		AccessLevel:      accessLevel,
		RequiresApproval: req.RequiresApproval,
		SortOrder:        req.SortOrder,
	}

	if err := s.ttRepo.Create(ctx, tt); err != nil {
		return nil, err
	}

	s.audit(ctx, &actorID, "TICKET_TYPE_CREATED", tt.ID.String())
	return toTicketTypeResponse(tt), nil
}

func (s *ticketTypeService) List(ctx context.Context, eventID uuid.UUID, viewerID *uuid.UUID) ([]responses.TicketTypeResponse, error) {
	includeHidden := false
	if viewerID != nil {
		event, _ := s.eventRepo.FindByID(ctx, eventID)
		if event != nil && s.canManage(ctx, event, *viewerID) {
			includeHidden = true
		}
	}
	list, err := s.ttRepo.FindByEvent(ctx, eventID, includeHidden)
	if err != nil {
		return nil, err
	}
	out := make([]responses.TicketTypeResponse, 0, len(list))
	for i := range list {
		out = append(out, *toTicketTypeResponse(&list[i]))
	}
	return out, nil
}

func (s *ticketTypeService) Update(ctx context.Context, actorID, eventID, ttID uuid.UUID, req *requests.TicketTypeUpdateRequest) (*responses.TicketTypeResponse, error) {
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
	tt, err := s.ttRepo.FindByID(ctx, ttID)
	if err != nil {
		return nil, err
	}
	if tt == nil || tt.EventID != eventID {
		return nil, fmt.Errorf("ticket type not found")
	}

	if req.Name != nil {
		tt.Name = *req.Name
	}
	if req.Description != nil {
		tt.Description = *req.Description
	}
	if req.ImageURL != nil {
		tt.ImageURL = *req.ImageURL
	}
	if req.PriceMinor != nil {
		tt.PriceMinor = *req.PriceMinor
	}
	if req.QuantityTotal != nil {
		if *req.QuantityTotal < tt.QuantitySold+tt.QuantityReserved {
			return nil, fmt.Errorf("cannot reduce total below sold + reserved (%d)", tt.QuantitySold+tt.QuantityReserved)
		}
		tt.QuantityTotal = *req.QuantityTotal
	}
	if req.PerUserLimit != nil {
		tt.PerUserLimit = *req.PerUserLimit
	}
	if req.SalesStartAt != nil {
		tt.SalesStartAt = req.SalesStartAt
	}
	if req.SalesEndAt != nil {
		tt.SalesEndAt = req.SalesEndAt
	}
	if req.IsHidden != nil {
		tt.IsHidden = *req.IsHidden
	}
	if req.IsActive != nil {
		tt.IsActive = *req.IsActive
	}
	if req.AccessLevel != nil {
		tt.AccessLevel = *req.AccessLevel
	}
	if req.RequiresApproval != nil {
		tt.RequiresApproval = *req.RequiresApproval
	}
	if req.SortOrder != nil {
		tt.SortOrder = *req.SortOrder
	}

	if err := s.ttRepo.Update(ctx, tt); err != nil {
		return nil, err
	}
	s.audit(ctx, &actorID, "TICKET_TYPE_UPDATED", tt.ID.String())
	return toTicketTypeResponse(tt), nil
}

func (s *ticketTypeService) Delete(ctx context.Context, actorID, eventID, ttID uuid.UUID) error {
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
	tt, err := s.ttRepo.FindByID(ctx, ttID)
	if err != nil {
		return err
	}
	if tt == nil || tt.EventID != eventID {
		return fmt.Errorf("ticket type not found")
	}
	if tt.QuantitySold > 0 {
		return fmt.Errorf("cannot delete a ticket type that has sold tickets")
	}
	if err := s.ttRepo.Delete(ctx, ttID); err != nil {
		return err
	}
	s.audit(ctx, &actorID, "TICKET_TYPE_DELETED", ttID.String())
	return nil
}

func (s *ticketTypeService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "ticket_type",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log: ", err)
	}
}

func toTicketTypeResponse(tt *models.TicketType) *responses.TicketTypeResponse {
	now := time.Now().UTC()
	return &responses.TicketTypeResponse{
		ID:                tt.ID,
		EventID:           tt.EventID,
		OccurrenceID:      tt.OccurrenceID,
		Name:              tt.Name,
		Description:       tt.Description,
		ImageURL:          tt.ImageURL,
		PriceMinor:        tt.PriceMinor,
		Currency:          tt.Currency,
		QuantityTotal:     tt.QuantityTotal,
		QuantitySold:      tt.QuantitySold,
		QuantityReserved:  tt.QuantityReserved,
		QuantityRemaining: tt.Remaining(),
		PerUserLimit:      tt.PerUserLimit,
		SalesStartAt:      tt.SalesStartAt,
		SalesEndAt:        tt.SalesEndAt,
		IsHidden:          tt.IsHidden,
		IsActive:          tt.IsActive,
		AccessLevel:       tt.AccessLevel,
		RequiresApproval:  tt.RequiresApproval,
		SortOrder:         tt.SortOrder,
		OnSale:            tt.IsOnSale(now),
		CreatedAt:         tt.CreatedAt,
		UpdatedAt:         tt.UpdatedAt,
	}
}

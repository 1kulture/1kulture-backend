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

type partnershipMetricService struct {
	metricRepo      repoInterfaces.PartnershipMetricRepository
	requestRepo     repoInterfaces.PartnershipRequestRepository
	coOrgRepo       repoInterfaces.EventCoOrganizerRepository
	notificationSvc serviceInterfaces.NotificationService
	auditLogRepo    repoInterfaces.AuditLogRepository
}

func NewPartnershipMetricService(
	metricRepo repoInterfaces.PartnershipMetricRepository,
	requestRepo repoInterfaces.PartnershipRequestRepository,
	coOrgRepo repoInterfaces.EventCoOrganizerRepository,
	notificationSvc serviceInterfaces.NotificationService,
	auditLogRepo repoInterfaces.AuditLogRepository,
) serviceInterfaces.PartnershipMetricService {
	return &partnershipMetricService{
		metricRepo:      metricRepo,
		requestRepo:     requestRepo,
		coOrgRepo:       coOrgRepo,
		notificationSvc: notificationSvc,
		auditLogRepo:    auditLogRepo,
	}
}

func (s *partnershipMetricService) authorizeForPartnership(ctx context.Context, actorID uuid.UUID, pr *models.PartnershipRequest) (string, error) {
	if pr.BrandUserID == actorID {
		return "brand", nil
	}
	if pr.OrganizerID == actorID {
		return "organizer", nil
	}
	isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, pr.EventID, actorID)
	if isCo {
		return "organizer", nil
	}
	return "", fmt.Errorf("forbidden")
}

func (s *partnershipMetricService) Add(ctx context.Context, actorID, partnershipID uuid.UUID, req *requests.PartnershipMetricRequest) (*responses.PartnershipMetricResponse, error) {
	pr, err := s.requestRepo.FindByID(ctx, partnershipID)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, fmt.Errorf("partnership not found")
	}
	if !pr.IsActiveState() && pr.Status != models.PartnershipStatusCompleted {
		return nil, fmt.Errorf("metrics can only be reported for accepted, active, or completed partnerships")
	}
	role, err := s.authorizeForPartnership(ctx, actorID, pr)
	if err != nil {
		return nil, err
	}

	m := &models.PartnershipMetric{
		PartnershipID: partnershipID,
		Key:           req.Key,
		Value:         req.Value,
		Currency:      req.Currency,
		Note:          req.Note,
		ReportedByID:  actorID,
		ReportedRole:  role,
	}
	if err := s.metricRepo.Create(ctx, m); err != nil {
		return nil, err
	}

	// Notify the other side
	var notifyUserID uuid.UUID
	if role == "brand" {
		notifyUserID = pr.OrganizerID
	} else {
		notifyUserID = pr.BrandUserID
	}
	_ = s.notificationSvc.Emit(
		ctx, notifyUserID,
		models.NotificationPartnershipMetricAdded,
		"New partnership metric reported",
		fmt.Sprintf("%s reported %s: %d", role, req.Key, req.Value),
		"partnership", pr.ID,
		map[string]interface{}{"metric_key": req.Key},
	)

	s.audit(ctx, &actorID, "PARTNERSHIP_METRIC_ADDED", m.ID.String())
	return toPartnershipMetricResponse(m), nil
}

func (s *partnershipMetricService) List(ctx context.Context, viewerID, partnershipID uuid.UUID) ([]responses.PartnershipMetricResponse, error) {
	pr, err := s.requestRepo.FindByID(ctx, partnershipID)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, fmt.Errorf("partnership not found")
	}
	if _, err := s.authorizeForPartnership(ctx, viewerID, pr); err != nil {
		return nil, err
	}
	list, err := s.metricRepo.FindByPartnership(ctx, partnershipID)
	if err != nil {
		return nil, err
	}
	out := make([]responses.PartnershipMetricResponse, 0, len(list))
	for i := range list {
		out = append(out, *toPartnershipMetricResponse(&list[i]))
	}
	return out, nil
}

func (s *partnershipMetricService) Delete(ctx context.Context, actorID, metricID uuid.UUID) error {
	m, err := s.metricRepo.FindByID(ctx, metricID)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("metric not found")
	}
	if m.ReportedByID != actorID {
		return fmt.Errorf("forbidden")
	}
	if err := s.metricRepo.Delete(ctx, metricID); err != nil {
		return err
	}
	s.audit(ctx, &actorID, "PARTNERSHIP_METRIC_DELETED", metricID.String())
	return nil
}

func (s *partnershipMetricService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "partnership_metric",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log: ", err)
	}
}

func toPartnershipMetricResponse(m *models.PartnershipMetric) *responses.PartnershipMetricResponse {
	return &responses.PartnershipMetricResponse{
		ID:            m.ID,
		PartnershipID: m.PartnershipID,
		Key:           m.Key,
		Value:         m.Value,
		Currency:      m.Currency,
		Note:          m.Note,
		ReportedByID:  m.ReportedByID,
		ReportedRole:  m.ReportedRole,
		CreatedAt:     m.CreatedAt,
	}
}

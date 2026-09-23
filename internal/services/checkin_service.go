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

type checkInService struct {
	checkinRepo  repoInterfaces.CheckInRepository
	ticketRepo   repoInterfaces.TicketRepository
	eventRepo    repoInterfaces.EventRepository
	coOrgRepo    repoInterfaces.EventCoOrganizerRepository
	staffRepo    repoInterfaces.EventStaffRepository
	auditLogRepo repoInterfaces.AuditLogRepository
}

func NewCheckInService(
	checkinRepo repoInterfaces.CheckInRepository,
	ticketRepo repoInterfaces.TicketRepository,
	eventRepo repoInterfaces.EventRepository,
	coOrgRepo repoInterfaces.EventCoOrganizerRepository,
	staffRepo repoInterfaces.EventStaffRepository,
	auditLogRepo repoInterfaces.AuditLogRepository,
) serviceInterfaces.CheckInService {
	return &checkInService{
		checkinRepo:  checkinRepo,
		ticketRepo:   ticketRepo,
		eventRepo:    eventRepo,
		coOrgRepo:    coOrgRepo,
		staffRepo:    staffRepo,
		auditLogRepo: auditLogRepo,
	}
}

func (s *checkInService) canScan(ctx context.Context, event *models.Event, userID uuid.UUID) bool {
	if event.OrganizerID == userID {
		return true
	}
	isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, event.ID, userID)
	if isCo {
		return true
	}
	isStaff, _ := s.staffRepo.IsStaff(ctx, event.ID, userID)
	return isStaff
}

func (s *checkInService) ScanByQR(ctx context.Context, actorID, eventID uuid.UUID, req *requests.CheckInRequest, ip string) (*responses.CheckInResponse, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	if !s.canScan(ctx, event, actorID) {
		return nil, fmt.Errorf("forbidden")
	}

	ticket, err := s.ticketRepo.FindByCode(ctx, req.QRPayload)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, fmt.Errorf("invalid ticket code")
	}
	if ticket.EventID != eventID {
		return nil, fmt.Errorf("ticket belongs to a different event")
	}

	// Idempotency — was it already scanned?
	if ticket.Status == models.TicketStatusUsed {
		existing, _ := s.checkinRepo.FindByTicket(ctx, ticket.ID)
		scannedAt := time.Now().UTC()
		if existing != nil {
			scannedAt = existing.ScannedAt
		}
		return &responses.CheckInResponse{
			Ticket:         *toTicketResponse(ticket),
			CheckedInAt:    scannedAt,
			ScannedBy:      actorID,
			AlreadyScanned: true,
		}, nil
	}

	if !ticket.IsUsable() {
		return nil, fmt.Errorf("ticket is not usable (status: %s)", ticket.Status)
	}

	now := time.Now().UTC()
	ci := &models.CheckIn{
		TicketID:   ticket.ID,
		EventID:    eventID,
		ScannedBy:  actorID,
		Method:     models.CheckInMethodQR,
		ScannedAt:  now,
		DeviceInfo: req.DeviceInfo,
		Location:   req.Location,
		IPAddress:  ip,
	}
	if err := s.checkinRepo.Create(ctx, ci); err != nil {
		return nil, err
	}

	ticket.Status = models.TicketStatusUsed
	ticket.UsedAt = &now
	ticket.ScannedBy = &actorID
	ticket.ScanLocation = req.Location
	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	s.audit(ctx, &actorID, "TICKET_CHECKED_IN", ticket.ID.String())
	return &responses.CheckInResponse{
		Ticket:         *toTicketResponse(ticket),
		CheckedInAt:    now,
		ScannedBy:      actorID,
		AlreadyScanned: false,
	}, nil
}

func (s *checkInService) ManualCheckIn(ctx context.Context, actorID, eventID uuid.UUID, req *requests.ManualCheckInRequest, ip string) (*responses.CheckInResponse, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	if !s.canScan(ctx, event, actorID) {
		return nil, fmt.Errorf("forbidden")
	}

	ticket, err := s.ticketRepo.FindByCode(ctx, req.TicketCode)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, fmt.Errorf("invalid ticket code")
	}
	if ticket.EventID != eventID {
		return nil, fmt.Errorf("ticket belongs to a different event")
	}
	if ticket.Status == models.TicketStatusUsed {
		existing, _ := s.checkinRepo.FindByTicket(ctx, ticket.ID)
		scannedAt := time.Now().UTC()
		if existing != nil {
			scannedAt = existing.ScannedAt
		}
		return &responses.CheckInResponse{
			Ticket:         *toTicketResponse(ticket),
			CheckedInAt:    scannedAt,
			ScannedBy:      actorID,
			AlreadyScanned: true,
		}, nil
	}
	if !ticket.IsUsable() {
		return nil, fmt.Errorf("ticket is not usable (status: %s)", ticket.Status)
	}

	now := time.Now().UTC()
	ci := &models.CheckIn{
		TicketID:   ticket.ID,
		EventID:    eventID,
		ScannedBy:  actorID,
		Method:     models.CheckInMethodManual,
		ScannedAt:  now,
		DeviceInfo: req.DeviceInfo,
		Location:   req.Location,
		IPAddress:  ip,
	}
	if err := s.checkinRepo.Create(ctx, ci); err != nil {
		return nil, err
	}

	ticket.Status = models.TicketStatusUsed
	ticket.UsedAt = &now
	ticket.ScannedBy = &actorID
	ticket.ScanLocation = req.Location
	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	s.audit(ctx, &actorID, "TICKET_MANUAL_CHECK_IN", ticket.ID.String())
	return &responses.CheckInResponse{
		Ticket:         *toTicketResponse(ticket),
		CheckedInAt:    now,
		ScannedBy:      actorID,
		AlreadyScanned: false,
	}, nil
}

func (s *checkInService) ListCheckIns(ctx context.Context, actorID, eventID uuid.UUID, page, perPage int) ([]responses.CheckInRecordResponse, int64, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, 0, err
	}
	if event == nil {
		return nil, 0, fmt.Errorf("event not found")
	}
	if !s.canScan(ctx, event, actorID) {
		return nil, 0, fmt.Errorf("forbidden")
	}
	list, total, err := s.checkinRepo.FindByEvent(ctx, eventID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.CheckInRecordResponse, 0, len(list))
	for i := range list {
		out = append(out, responses.CheckInRecordResponse{
			ID:        list[i].ID,
			TicketID:  list[i].TicketID,
			EventID:   list[i].EventID,
			ScannedBy: list[i].ScannedBy,
			Method:    string(list[i].Method),
			ScannedAt: list[i].ScannedAt,
			Location:  list[i].Location,
		})
	}
	return out, total, nil
}

func (s *checkInService) Stats(ctx context.Context, actorID, eventID uuid.UUID) (*responses.CheckInStatsResponse, error) {
	event, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}
	if !s.canScan(ctx, event, actorID) {
		return nil, fmt.Errorf("forbidden")
	}
	total, err := s.ticketRepo.CountByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	checkedIn, err := s.ticketRepo.CountUsedByEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}
	percent := 0.0
	if total > 0 {
		percent = float64(checkedIn) / float64(total) * 100
	}
	return &responses.CheckInStatsResponse{
		EventID:          eventID,
		TotalTickets:     int(total),
		CheckedIn:        int(checkedIn),
		Remaining:        int(total - checkedIn),
		CheckedInPercent: percent,
	}, nil
}

func (s *checkInService) audit(ctx context.Context, actorID *uuid.UUID, action, resourceID string) {
	log := &models.AuditLog{
		UserID:     actorID,
		Action:     action,
		Resource:   "check_in",
		ResourceID: resourceID,
		IPAddress:  getContextString(ctx, "ip_address"),
		UserAgent:  getContextString(ctx, "user_agent"),
		Status:     "success",
	}
	if err := s.auditLogRepo.Create(ctx, log); err != nil {
		logger.Error("Failed to write audit log: ", err)
	}
}

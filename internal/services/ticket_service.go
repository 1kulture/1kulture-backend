package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
)

type ticketService struct {
	ticketRepo repoInterfaces.TicketRepository
	eventRepo  repoInterfaces.EventRepository
	coOrgRepo  repoInterfaces.EventCoOrganizerRepository
}

func NewTicketService(
	ticketRepo repoInterfaces.TicketRepository,
	eventRepo repoInterfaces.EventRepository,
	coOrgRepo repoInterfaces.EventCoOrganizerRepository,
) serviceInterfaces.TicketService {
	return &ticketService{
		ticketRepo: ticketRepo,
		eventRepo:  eventRepo,
		coOrgRepo:  coOrgRepo,
	}
}

func (s *ticketService) GetTicket(ctx context.Context, viewerID, ticketID uuid.UUID) (*responses.TicketWithQRResponse, error) {
	t, err := s.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, fmt.Errorf("ticket not found")
	}
	if !s.canView(ctx, t, viewerID) {
		return nil, fmt.Errorf("forbidden")
	}
	return &responses.TicketWithQRResponse{
		TicketResponse: *toTicketResponse(t),
		QRPayload:      t.QRPayload,
	}, nil
}

func (s *ticketService) GetTicketByCode(ctx context.Context, viewerID uuid.UUID, code string) (*responses.TicketWithQRResponse, error) {
	t, err := s.ticketRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, fmt.Errorf("ticket not found")
	}
	if !s.canView(ctx, t, viewerID) {
		return nil, fmt.Errorf("forbidden")
	}
	return &responses.TicketWithQRResponse{
		TicketResponse: *toTicketResponse(t),
		QRPayload:      t.QRPayload,
	}, nil
}

func (s *ticketService) ListMyTickets(ctx context.Context, userID uuid.UUID, page, perPage int) ([]responses.TicketResponse, int64, error) {
	list, total, err := s.ticketRepo.FindByUser(ctx, userID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.TicketResponse, 0, len(list))
	for i := range list {
		out = append(out, *toTicketResponse(&list[i]))
	}
	return out, total, nil
}

func (s *ticketService) ListEventTickets(ctx context.Context, actorID, eventID uuid.UUID, page, perPage int) ([]responses.TicketResponse, int64, error) {
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
	list, total, err := s.ticketRepo.FindByEvent(ctx, eventID, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.TicketResponse, 0, len(list))
	for i := range list {
		out = append(out, *toTicketResponse(&list[i]))
	}
	return out, total, nil
}

func (s *ticketService) canView(ctx context.Context, t *models.Ticket, viewerID uuid.UUID) bool {
	if t.UserID == viewerID {
		return true
	}
	event, _ := s.eventRepo.FindByID(ctx, t.EventID)
	if event == nil {
		return false
	}
	if event.OrganizerID == viewerID {
		return true
	}
	isCo, _ := s.coOrgRepo.IsCoOrganizer(ctx, t.EventID, viewerID)
	return isCo
}

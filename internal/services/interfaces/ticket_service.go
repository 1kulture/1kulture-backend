package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type TicketService interface {
	GetTicket(ctx context.Context, viewerID, ticketID uuid.UUID) (*responses.TicketWithQRResponse, error)
	GetTicketByCode(ctx context.Context, viewerID uuid.UUID, code string) (*responses.TicketWithQRResponse, error)
	ListMyTickets(ctx context.Context, userID uuid.UUID, page, perPage int) ([]responses.TicketResponse, int64, error)
	ListEventTickets(ctx context.Context, actorID, eventID uuid.UUID, page, perPage int) ([]responses.TicketResponse, int64, error)
}

package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type TicketTypeService interface {
	Create(ctx context.Context, actorID, eventID uuid.UUID, req *requests.TicketTypeCreateRequest) (*responses.TicketTypeResponse, error)
	List(ctx context.Context, eventID uuid.UUID, viewerID *uuid.UUID) ([]responses.TicketTypeResponse, error)
	Update(ctx context.Context, actorID, eventID, ttID uuid.UUID, req *requests.TicketTypeUpdateRequest) (*responses.TicketTypeResponse, error)
	Delete(ctx context.Context, actorID, eventID, ttID uuid.UUID) error
}

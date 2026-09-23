package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type PartnershipService interface {
	// Brand side
	CreateRequest(ctx context.Context, brandUserID, eventID uuid.UUID, req *requests.PartnershipRequestCreate) (*responses.PartnershipRequestResponse, error)
	ListMyBrandRequests(ctx context.Context, brandUserID uuid.UUID, status string, page, perPage int) ([]responses.PartnershipRequestResponse, int64, error)
	CancelRequest(ctx context.Context, brandUserID, requestID uuid.UUID, req *requests.PartnershipDeclineRequest) (*responses.PartnershipRequestResponse, error)

	// Organizer side
	ListEventRequests(ctx context.Context, actorID, eventID uuid.UUID, status string, page, perPage int) ([]responses.PartnershipRequestResponse, int64, error)
	DecideRequest(ctx context.Context, actorID, requestID uuid.UUID, accept bool, reason string) (*responses.PartnershipRequestResponse, error)
	MarkActivated(ctx context.Context, actorID, requestID uuid.UUID) (*responses.PartnershipRequestResponse, error)
	MarkCompleted(ctx context.Context, actorID, requestID uuid.UUID) (*responses.PartnershipRequestResponse, error)

	// Shared
	GetRequest(ctx context.Context, viewerID, requestID uuid.UUID) (*responses.PartnershipRequestResponse, error)
	ListAll(ctx context.Context, filter interfaces.PartnershipRequestFilter, viewerID uuid.UUID) ([]responses.PartnershipRequestResponse, int64, error)
	Dashboard(ctx context.Context, viewerID uuid.UUID, eventID *uuid.UUID) (*responses.PartnershipDashboardResponse, error)
}

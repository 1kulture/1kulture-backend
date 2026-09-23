package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type PartnershipConfigService interface {
	// UpsertConfig creates or updates the partnership config for an event.
	UpsertConfig(ctx context.Context, actorID, eventID uuid.UUID, req *requests.EventPartnershipConfigRequest) (*responses.EventPartnershipConfigResponse, error)

	// GetConfig returns the config (public if event is published).
	GetConfig(ctx context.Context, eventID uuid.UUID, viewerID *uuid.UUID) (*responses.EventPartnershipConfigResponse, error)

	// Opportunity CRUD (organizer / co-organizer only)
	AddOpportunity(ctx context.Context, actorID, eventID uuid.UUID, req *requests.OpportunityCreateRequest) (*responses.OpportunityResponse, error)
	UpdateOpportunity(ctx context.Context, actorID, eventID, oppID uuid.UUID, req *requests.OpportunityUpdateRequest) (*responses.OpportunityResponse, error)
	DeleteOpportunity(ctx context.Context, actorID, eventID, oppID uuid.UUID) error
	ListOpportunities(ctx context.Context, eventID uuid.UUID, onlyActive bool) ([]responses.OpportunityResponse, error)
}

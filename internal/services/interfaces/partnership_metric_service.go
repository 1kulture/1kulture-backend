package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type PartnershipMetricService interface {
	Add(ctx context.Context, actorID, partnershipID uuid.UUID, req *requests.PartnershipMetricRequest) (*responses.PartnershipMetricResponse, error)
	List(ctx context.Context, viewerID, partnershipID uuid.UUID) ([]responses.PartnershipMetricResponse, error)
	Delete(ctx context.Context, actorID, metricID uuid.UUID) error
}

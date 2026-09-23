package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type AffiliateCodeService interface {
	Create(ctx context.Context, actorID, partnershipID uuid.UUID, req *requests.CreateAffiliateCodeRequest) (*responses.AffiliateCodeResponse, error)
	List(ctx context.Context, viewerID, partnershipID uuid.UUID) ([]responses.AffiliateCodeResponse, error)
}

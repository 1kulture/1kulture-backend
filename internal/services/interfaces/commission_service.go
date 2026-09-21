package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type CommissionService interface {
	CreateTier(ctx context.Context, req *requests.CreateCommissionTierRequest) (*responses.CommissionTierResponse, error)
	ListTiers(ctx context.Context, onlyActive bool) ([]responses.CommissionTierResponse, error)
	UpdateTier(ctx context.Context, id uuid.UUID, req *requests.UpdateCommissionTierRequest) (*responses.CommissionTierResponse, error)
	DeleteTier(ctx context.Context, id uuid.UUID) error

	SetOverride(ctx context.Context, actorID uuid.UUID, req *requests.SetOrganizerCommissionOverrideRequest) (*responses.OrganizerCommissionOverrideResponse, error)
	GetOverride(ctx context.Context, organizerID uuid.UUID) (*responses.OrganizerCommissionOverrideResponse, error)
	DeleteOverride(ctx context.Context, organizerID uuid.UUID) error

	// Resolve returns the effective commission rate (bps) for an organizer.
	// Resolution order: override → tier by lifetime revenue → default rate from settings.
	ResolveRateBps(ctx context.Context, organizerID uuid.UUID) (int, error)
}

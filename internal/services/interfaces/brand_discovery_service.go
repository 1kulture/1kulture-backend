package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type BrandDiscoveryFilter struct {
	Category        string
	City            string
	Country         string
	PartnershipType string
	MinAudience     int
	MaxAudience     int
	StartFromISO    string
	StartToISO      string
	Search          string
	Page            int
	PerPage         int
}

type BrandDiscoveryService interface {
	// DiscoverEvents returns events with partnership mode enabled, filtered.
	DiscoverEvents(ctx context.Context, brandUserID uuid.UUID, filter BrandDiscoveryFilter) ([]responses.EventSummaryResponse, int64, error)

	// GetEventDetail returns the event + partnership config + opportunities for a brand.
	GetEventDetail(ctx context.Context, brandUserID, eventID uuid.UUID) (*responses.EventResponse, *responses.EventPartnershipConfigResponse, error)

	// Recommended returns a lightweight list of events matching the brand's preferred categories.
	Recommended(ctx context.Context, brandUserID uuid.UUID, limit int) ([]responses.EventSummaryResponse, error)

	// Dashboard returns brand-side stats.
	Dashboard(ctx context.Context, brandUserID uuid.UUID) (*responses.BrandDashboardResponse, error)
}

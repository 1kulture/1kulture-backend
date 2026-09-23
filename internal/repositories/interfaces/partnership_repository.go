package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type EventPartnershipConfigRepository interface {
	Create(ctx context.Context, cfg *models.EventPartnershipConfig) error
	FindByEventID(ctx context.Context, eventID uuid.UUID) (*models.EventPartnershipConfig, error)
	Update(ctx context.Context, cfg *models.EventPartnershipConfig) error
	Delete(ctx context.Context, eventID uuid.UUID) error
}

type PartnershipOpportunityRepository interface {
	Create(ctx context.Context, opp *models.PartnershipOpportunity) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.PartnershipOpportunity, error)
	FindByEvent(ctx context.Context, eventID uuid.UUID, onlyActive bool) ([]models.PartnershipOpportunity, error)
	Update(ctx context.Context, opp *models.PartnershipOpportunity) error
	Delete(ctx context.Context, id uuid.UUID) error
	DecrementSlots(ctx context.Context, id uuid.UUID, delta int) error
	IncrementSlots(ctx context.Context, id uuid.UUID, delta int) error
}

type PartnershipRequestFilter struct {
	BrandProfileID *uuid.UUID
	BrandUserID    *uuid.UUID
	EventID        *uuid.UUID
	OrganizerID    *uuid.UUID
	Status         string
	Statuses       []string
	Page           int
	PerPage        int
}

type PartnershipRequestRepository interface {
	Create(ctx context.Context, req *models.PartnershipRequest) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.PartnershipRequest, error)
	FindWithRelations(ctx context.Context, id uuid.UUID) (*models.PartnershipRequest, error)
	Update(ctx context.Context, req *models.PartnershipRequest) error
	List(ctx context.Context, filter PartnershipRequestFilter) ([]models.PartnershipRequest, int64, error)
	CountByStatus(ctx context.Context, filter PartnershipRequestFilter) (map[string]int, error)
	ExistsActive(ctx context.Context, brandUserID, eventID uuid.UUID) (bool, error)
}

type PartnershipMetricRepository interface {
	Create(ctx context.Context, m *models.PartnershipMetric) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.PartnershipMetric, error)
	FindByPartnership(ctx context.Context, partnershipID uuid.UUID) ([]models.PartnershipMetric, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type AffiliateCodeRepository interface {
	Create(ctx context.Context, a *models.AffiliateCode) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.AffiliateCode, error)
	FindByCode(ctx context.Context, code string) (*models.AffiliateCode, error)
	FindByPartnership(ctx context.Context, partnershipID uuid.UUID) ([]models.AffiliateCode, error)
	Update(ctx context.Context, a *models.AffiliateCode) error
	IncrementClicks(ctx context.Context, id uuid.UUID, delta int64) error
	IncrementConversions(ctx context.Context, id uuid.UUID, revenueMinor, commissionMinor int64) error
}

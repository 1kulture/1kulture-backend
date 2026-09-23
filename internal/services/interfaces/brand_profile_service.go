package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type BrandProfileService interface {
	CreateForUser(ctx context.Context, userID uuid.UUID, req *requests.CreateBrandProfileRequest) (*responses.BrandProfileResponse, error)
	GetMine(ctx context.Context, userID uuid.UUID) (*responses.BrandProfileResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*responses.BrandProfileResponse, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*responses.BrandProfileResponse, error)
	Update(ctx context.Context, userID uuid.UUID, req *requests.UpdateBrandProfileRequest) (*responses.BrandProfileResponse, error)
	List(ctx context.Context, industry, location string, page, perPage int) ([]responses.BrandProfileResponse, int64, error)
}

package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/requests"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*responses.UserResponse, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *requests.UpdateProfileRequest) (*responses.UserResponse, error)
	UpdateRole(ctx context.Context, userID uuid.UUID, req *requests.UpdateRoleRequest) error
	SubmitKYC(ctx context.Context, userID uuid.UUID, req *requests.KYCRequest) (*responses.KYCResponse, error)
	GetKYCStatus(ctx context.Context, userID uuid.UUID) (*responses.KYCResponse, error)
}

package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type KYCRepository interface {
	Create(ctx context.Context, kyc *models.KYCVerification) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.KYCVerification, error)
	FindByUserAndRole(ctx context.Context, userID, roleID uuid.UUID) (*models.KYCVerification, error)
	FindLatestByUser(ctx context.Context, userID uuid.UUID) (*models.KYCVerification, error)
	Update(ctx context.Context, kyc *models.KYCVerification) error
	List(ctx context.Context, page, perPage int) ([]models.KYCVerification, int64, error)
	ListByStatus(ctx context.Context, status models.KYCStatus, page, perPage int) ([]models.KYCVerification, int64, error)
}

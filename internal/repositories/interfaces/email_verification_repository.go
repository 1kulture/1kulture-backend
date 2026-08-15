package interfaces

import (
	"context"
	"time"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type EmailVerificationRepository interface {
	Create(ctx context.Context, verification *models.EmailVerification) error
	FindByEmailAndCode(ctx context.Context, email, code string) (*models.EmailVerification, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.EmailVerification, error)
	Update(ctx context.Context, verification *models.EmailVerification) error
	DeleteExpired(ctx context.Context) error
	MarkAsVerified(ctx context.Context, id uuid.UUID, verifiedAt time.Time) error
}

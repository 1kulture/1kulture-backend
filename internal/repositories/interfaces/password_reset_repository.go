package interfaces

import (
	"context"
	"time"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type PasswordResetRepository interface {
	Create(ctx context.Context, reset *models.PasswordReset) error
	FindByToken(ctx context.Context, token string) (*models.PasswordReset, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*models.PasswordReset, error)
	MarkAsUsed(ctx context.Context, id uuid.UUID, usedAt time.Time) error
	DeleteExpired(ctx context.Context) error
}

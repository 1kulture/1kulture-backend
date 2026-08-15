package interfaces

import (
	"context"
	"time"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, perPage int) ([]models.User, int64, error)
	UpdateEmailVerification(ctx context.Context, userID uuid.UUID, verifiedAt time.Time) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, lastLoginAt time.Time) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UpdateStatus(ctx context.Context, userID uuid.UUID, status models.UserStatus) error
	WithRoles(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

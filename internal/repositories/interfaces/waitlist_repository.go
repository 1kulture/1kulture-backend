package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
)

type WaitlistRepository interface {
	Create(ctx context.Context, entry *models.WaitlistEntry) error
	FindByEmail(ctx context.Context, email string) ([]models.WaitlistEntry, error)
}

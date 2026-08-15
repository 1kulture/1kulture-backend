package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error)
	List(ctx context.Context, page, perPage int) ([]models.AuditLog, int64, error)
	ListByUser(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.AuditLog, int64, error)
	ListByAction(ctx context.Context, action string, page, perPage int) ([]models.AuditLog, int64, error)
}

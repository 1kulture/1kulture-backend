package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/google/uuid"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *models.Notification) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Notification, error)
	ListByUser(ctx context.Context, userID uuid.UUID, unreadOnly bool, page, perPage int) ([]models.Notification, int64, error)
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

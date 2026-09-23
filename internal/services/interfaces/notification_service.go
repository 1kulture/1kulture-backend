package interfaces

import (
	"context"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/responses"
	"github.com/google/uuid"
)

type NotificationService interface {
	// Emit creates a notification for a user.
	Emit(ctx context.Context, userID uuid.UUID, notifType models.NotificationType, title, body string, refType string, refID uuid.UUID, data map[string]interface{}) error

	List(ctx context.Context, userID uuid.UUID, unreadOnly bool, page, perPage int) ([]responses.NotificationResponse, int64, error)
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkRead(ctx context.Context, userID, notificationID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, userID, notificationID uuid.UUID) error
}

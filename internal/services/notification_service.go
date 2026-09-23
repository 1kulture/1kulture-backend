package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
	"github.com/1kulture/1kulture-backend/internal/responses"
	serviceInterfaces "github.com/1kulture/1kulture-backend/internal/services/interfaces"
	"github.com/1kulture/1kulture-backend/internal/utils/logger"
)

type notificationService struct {
	notificationRepo repoInterfaces.NotificationRepository
}

func NewNotificationService(repo repoInterfaces.NotificationRepository) serviceInterfaces.NotificationService {
	return &notificationService{notificationRepo: repo}
}

func (s *notificationService) Emit(
	ctx context.Context,
	userID uuid.UUID,
	notifType models.NotificationType,
	title, body string,
	refType string,
	refID uuid.UUID,
	data map[string]interface{},
) error {
	var dataJSON datatypes.JSON
	if len(data) > 0 {
		b, err := jsonMarshal(data)
		if err == nil {
			dataJSON = datatypes.JSON(b)
		}
	}

	n := &models.Notification{
		UserID:  userID,
		Type:    notifType,
		Channel: models.NotificationChannelInApp,
		Title:   title,
		Body:    body,
		RefType: refType,
		RefID:   refID,
		Data:    dataJSON,
	}

	if err := s.notificationRepo.Create(ctx, n); err != nil {
		logger.Error("Failed to emit notification: ", err)
		return err
	}
	return nil
}

func (s *notificationService) List(ctx context.Context, userID uuid.UUID, unreadOnly bool, page, perPage int) ([]responses.NotificationResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	list, total, err := s.notificationRepo.ListByUser(ctx, userID, unreadOnly, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	out := make([]responses.NotificationResponse, 0, len(list))
	for i := range list {
		out = append(out, *toNotificationResponse(&list[i]))
	}
	return out, total, nil
}

func (s *notificationService) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notificationRepo.CountUnread(ctx, userID)
}

func (s *notificationService) MarkRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	n, err := s.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return err
	}
	if n == nil {
		return fmt.Errorf("notification not found")
	}
	if n.UserID != userID {
		return fmt.Errorf("forbidden")
	}
	return s.notificationRepo.MarkAsRead(ctx, notificationID)
}

func (s *notificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.notificationRepo.MarkAllRead(ctx, userID)
}

func (s *notificationService) Delete(ctx context.Context, userID, notificationID uuid.UUID) error {
	n, err := s.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return err
	}
	if n == nil {
		return nil
	}
	if n.UserID != userID {
		return fmt.Errorf("forbidden")
	}
	return s.notificationRepo.Delete(ctx, notificationID)
}

func toNotificationResponse(n *models.Notification) *responses.NotificationResponse {
	var data map[string]interface{}
	if len(n.Data) > 0 && string(n.Data) != "null" {
		_ = jsonUnmarshal(n.Data, &data)
	}
	return &responses.NotificationResponse{
		ID:        n.ID,
		UserID:    n.UserID,
		Type:      string(n.Type),
		Channel:   string(n.Channel),
		Title:     n.Title,
		Body:      n.Body,
		RefType:   n.RefType,
		RefID:     n.RefID,
		Data:      data,
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
	}
}

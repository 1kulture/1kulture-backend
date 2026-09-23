package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) repoInterfaces.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *models.Notification) error {
	if err := r.db.WithContext(ctx).Create(n).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

func (r *notificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	var n models.Notification
	if err := r.db.WithContext(ctx).First(&n, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find notification: %w", err)
	}
	return &n, nil
}

func (r *notificationRepository) ListByUser(ctx context.Context, userID uuid.UUID, unreadOnly bool, page, perPage int) ([]models.Notification, int64, error) {
	var list []models.Notification
	var total int64
	q := r.db.WithContext(ctx).Model(&models.Notification{}).Where("user_id = ?", userID)
	if unreadOnly {
		q = q.Where("read_at IS NULL")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Order("created_at desc").Limit(perPage).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}
	return list, total, nil
}

func (r *notificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}
	return count, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ? AND read_at IS NULL", id).
		Update("read_at", now).Error; err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

func (r *notificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", now).Error; err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}

func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.Notification{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}
	return nil
}

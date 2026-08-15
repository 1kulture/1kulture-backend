package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type passwordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) interfaces.PasswordResetRepository {
	return &passwordResetRepository{
		db: db,
	}
}

func (r *passwordResetRepository) Create(ctx context.Context, reset *models.PasswordReset) error {
	if err := r.db.WithContext(ctx).Create(reset).Error; err != nil {
		return fmt.Errorf("failed to create password reset: %w", err)
	}
	return nil
}

func (r *passwordResetRepository) FindByToken(ctx context.Context, token string) (*models.PasswordReset, error) {
	var reset models.PasswordReset
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&reset).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find password reset by token: %w", err)
	}
	return &reset, nil
}

func (r *passwordResetRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.PasswordReset, error) {
	var reset models.PasswordReset
	if err := r.db.WithContext(ctx).Where("user_id = ? AND used_at IS NULL", userID).Order("created_at DESC").First(&reset).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find password reset by user: %w", err)
	}
	return &reset, nil
}

func (r *passwordResetRepository) MarkAsUsed(ctx context.Context, id uuid.UUID, usedAt time.Time) error {
	if err := r.db.WithContext(ctx).Model(&models.PasswordReset{}).Where("id = ?", id).Update("used_at", usedAt).Error; err != nil {
		return fmt.Errorf("failed to mark password reset as used: %w", err)
	}
	return nil
}

func (r *passwordResetRepository) DeleteExpired(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.PasswordReset{}).Error; err != nil {
		return fmt.Errorf("failed to delete expired password resets: %w", err)
	}
	return nil
}

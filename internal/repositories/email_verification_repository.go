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

type emailVerificationRepository struct {
	db *gorm.DB
}

func NewEmailVerificationRepository(db *gorm.DB) interfaces.EmailVerificationRepository {
	return &emailVerificationRepository{
		db: db,
	}
}

func (r *emailVerificationRepository) Create(ctx context.Context, verification *models.EmailVerification) error {
	if err := r.db.WithContext(ctx).Create(verification).Error; err != nil {
		return fmt.Errorf("failed to create email verification: %w", err)
	}
	return nil
}

func (r *emailVerificationRepository) FindByEmailAndCode(ctx context.Context, email, code string) (*models.EmailVerification, error) {
	var verification models.EmailVerification
	if err := r.db.WithContext(ctx).
		Where("email = ? AND code = ?", email, code).
		Order("created_at DESC").
		First(&verification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find email verification: %w", err)
	}
	return &verification, nil
}

func (r *emailVerificationRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.EmailVerification, error) {
	var verification models.EmailVerification
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&verification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find email verification by user: %w", err)
	}
	return &verification, nil
}

func (r *emailVerificationRepository) Update(ctx context.Context, verification *models.EmailVerification) error {
	if err := r.db.WithContext(ctx).Save(verification).Error; err != nil {
		return fmt.Errorf("failed to update email verification: %w", err)
	}
	return nil
}

func (r *emailVerificationRepository) DeleteExpired(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.EmailVerification{}).Error; err != nil {
		return fmt.Errorf("failed to delete expired verifications: %w", err)
	}
	return nil
}

func (r *emailVerificationRepository) MarkAsVerified(ctx context.Context, id uuid.UUID, verifiedAt time.Time) error {
	if err := r.db.WithContext(ctx).Model(&models.EmailVerification{}).Where("id = ?", id).Update("verified_at", verifiedAt).Error; err != nil {
		return fmt.Errorf("failed to mark email as verified: %w", err)
	}
	return nil
}

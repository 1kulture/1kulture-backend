package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

// ==========================================================
// PromoCodeRepository
// ==========================================================

type promoCodeRepository struct {
	db *gorm.DB
}

func NewPromoCodeRepository(db *gorm.DB) repoInterfaces.PromoCodeRepository {
	return &promoCodeRepository{db: db}
}

func (r *promoCodeRepository) Create(ctx context.Context, pc *models.PromoCode) error {
	if err := r.db.WithContext(ctx).Create(pc).Error; err != nil {
		return fmt.Errorf("failed to create promo code: %w", err)
	}
	return nil
}

func (r *promoCodeRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.PromoCode, error) {
	var pc models.PromoCode
	if err := r.db.WithContext(ctx).First(&pc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find promo code: %w", err)
	}
	return &pc, nil
}

func (r *promoCodeRepository) FindByCode(ctx context.Context, code string) (*models.PromoCode, error) {
	var pc models.PromoCode
	if err := r.db.WithContext(ctx).First(&pc, "code = ?", code).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find promo code by code: %w", err)
	}
	return &pc, nil
}

func (r *promoCodeRepository) FindByEvent(ctx context.Context, eventID uuid.UUID) ([]models.PromoCode, error) {
	var list []models.PromoCode
	if err := r.db.WithContext(ctx).
		Where("event_id = ? OR event_id IS NULL", eventID).
		Order("created_at desc").
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to list promo codes: %w", err)
	}
	return list, nil
}

func (r *promoCodeRepository) Update(ctx context.Context, pc *models.PromoCode) error {
	if err := r.db.WithContext(ctx).Save(pc).Error; err != nil {
		return fmt.Errorf("failed to update promo code: %w", err)
	}
	return nil
}

func (r *promoCodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.PromoCode{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete promo code: %w", err)
	}
	return nil
}

func (r *promoCodeRepository) IncrementUsage(ctx context.Context, id uuid.UUID, delta int) error {
	if err := r.db.WithContext(ctx).Exec(`
		UPDATE promo_codes
		SET usage_count = GREATEST(usage_count + ?, 0),
		    updated_at = NOW()
		WHERE id = ?
	`, delta, id).Error; err != nil {
		return fmt.Errorf("failed to increment promo usage: %w", err)
	}
	return nil
}

func (r *promoCodeRepository) CountUserRedemptions(ctx context.Context, promoID, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.PromoCodeRedemption{}).
		Where("promo_code_id = ? AND user_id = ?", promoID, userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count user redemptions: %w", err)
	}
	return count, nil
}

// ==========================================================
// PromoCodeRedemptionRepository
// ==========================================================

type promoCodeRedemptionRepository struct {
	db *gorm.DB
}

func NewPromoCodeRedemptionRepository(db *gorm.DB) repoInterfaces.PromoCodeRedemptionRepository {
	return &promoCodeRedemptionRepository{db: db}
}

func (r *promoCodeRedemptionRepository) Create(ctx context.Context, red *models.PromoCodeRedemption) error {
	if err := r.db.WithContext(ctx).Create(red).Error; err != nil {
		return fmt.Errorf("failed to create promo redemption: %w", err)
	}
	return nil
}

func (r *promoCodeRedemptionRepository) FindByOrder(ctx context.Context, orderID uuid.UUID) (*models.PromoCodeRedemption, error) {
	var red models.PromoCodeRedemption
	if err := r.db.WithContext(ctx).First(&red, "order_id = ?", orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find redemption by order: %w", err)
	}
	return &red, nil
}

package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type affiliateCodeRepository struct {
	db *gorm.DB
}

func NewAffiliateCodeRepository(db *gorm.DB) repoInterfaces.AffiliateCodeRepository {
	return &affiliateCodeRepository{db: db}
}

func (r *affiliateCodeRepository) Create(ctx context.Context, a *models.AffiliateCode) error {
	if err := r.db.WithContext(ctx).Create(a).Error; err != nil {
		return fmt.Errorf("failed to create affiliate code: %w", err)
	}
	return nil
}

func (r *affiliateCodeRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.AffiliateCode, error) {
	var a models.AffiliateCode
	if err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find affiliate code: %w", err)
	}
	return &a, nil
}

func (r *affiliateCodeRepository) FindByCode(ctx context.Context, code string) (*models.AffiliateCode, error) {
	var a models.AffiliateCode
	if err := r.db.WithContext(ctx).First(&a, "code = ?", code).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find affiliate code by code: %w", err)
	}
	return &a, nil
}

func (r *affiliateCodeRepository) FindByPartnership(ctx context.Context, partnershipID uuid.UUID) ([]models.AffiliateCode, error) {
	var list []models.AffiliateCode
	if err := r.db.WithContext(ctx).
		Where("partnership_id = ?", partnershipID).
		Order("created_at desc").
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to list affiliate codes: %w", err)
	}
	return list, nil
}

func (r *affiliateCodeRepository) Update(ctx context.Context, a *models.AffiliateCode) error {
	if err := r.db.WithContext(ctx).Save(a).Error; err != nil {
		return fmt.Errorf("failed to update affiliate code: %w", err)
	}
	return nil
}

func (r *affiliateCodeRepository) IncrementClicks(ctx context.Context, id uuid.UUID, delta int64) error {
	if delta <= 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Exec(`
		UPDATE affiliate_codes
		SET clicks = clicks + ?,
		    updated_at = NOW()
		WHERE id = ?
	`, delta, id).Error; err != nil {
		return fmt.Errorf("failed to increment clicks: %w", err)
	}
	return nil
}

func (r *affiliateCodeRepository) IncrementConversions(ctx context.Context, id uuid.UUID, revenueMinor, commissionMinor int64) error {
	if err := r.db.WithContext(ctx).Exec(`
		UPDATE affiliate_codes
		SET conversions = conversions + 1,
		    revenue_minor = revenue_minor + ?,
		    commission_minor = commission_minor + ?,
		    updated_at = NOW()
		WHERE id = ?
	`, revenueMinor, commissionMinor, id).Error; err != nil {
		return fmt.Errorf("failed to increment conversions: %w", err)
	}
	return nil
}

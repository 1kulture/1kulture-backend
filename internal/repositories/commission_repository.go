package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type commissionTierRepository struct {
	db *gorm.DB
}

func NewCommissionTierRepository(db *gorm.DB) interfaces.CommissionTierRepository {
	return &commissionTierRepository{db: db}
}

func (r *commissionTierRepository) Create(ctx context.Context, tier *models.CommissionTier) error {
	if err := r.db.WithContext(ctx).Create(tier).Error; err != nil {
		return fmt.Errorf("failed to create commission tier: %w", err)
	}
	return nil
}

func (r *commissionTierRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.CommissionTier, error) {
	var t models.CommissionTier
	if err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find commission tier: %w", err)
	}
	return &t, nil
}

func (r *commissionTierRepository) FindAll(ctx context.Context, onlyActive bool) ([]models.CommissionTier, error) {
	var tiers []models.CommissionTier
	q := r.db.WithContext(ctx).Model(&models.CommissionTier{})
	if onlyActive {
		q = q.Where("is_active = ?", true)
	}
	if err := q.Order("priority asc, min_revenue_minor asc").Find(&tiers).Error; err != nil {
		return nil, fmt.Errorf("failed to list commission tiers: %w", err)
	}
	return tiers, nil
}

func (r *commissionTierRepository) Update(ctx context.Context, tier *models.CommissionTier) error {
	if err := r.db.WithContext(ctx).Save(tier).Error; err != nil {
		return fmt.Errorf("failed to update commission tier: %w", err)
	}
	return nil
}

func (r *commissionTierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.CommissionTier{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete commission tier: %w", err)
	}
	return nil
}

// ---------- Organizer Override ----------

type organizerCommissionOverrideRepository struct {
	db *gorm.DB
}

func NewOrganizerCommissionOverrideRepository(db *gorm.DB) interfaces.OrganizerCommissionOverrideRepository {
	return &organizerCommissionOverrideRepository{db: db}
}

func (r *organizerCommissionOverrideRepository) Upsert(ctx context.Context, override *models.OrganizerCommissionOverride) error {
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "organizer_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"rate_bps", "reason", "set_by", "updated_at"}),
		}).
		Create(override).Error; err != nil {
		return fmt.Errorf("failed to upsert organizer commission override: %w", err)
	}
	return nil
}

func (r *organizerCommissionOverrideRepository) FindByOrganizer(ctx context.Context, organizerID uuid.UUID) (*models.OrganizerCommissionOverride, error) {
	var o models.OrganizerCommissionOverride
	if err := r.db.WithContext(ctx).First(&o, "organizer_id = ?", organizerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find organizer commission override: %w", err)
	}
	return &o, nil
}

func (r *organizerCommissionOverrideRepository) Delete(ctx context.Context, organizerID uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("organizer_id = ?", organizerID).Delete(&models.OrganizerCommissionOverride{}).Error; err != nil {
		return fmt.Errorf("failed to delete organizer commission override: %w", err)
	}
	return nil
}

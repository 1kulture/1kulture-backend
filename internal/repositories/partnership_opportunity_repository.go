package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type partnershipOpportunityRepository struct {
	db *gorm.DB
}

func NewPartnershipOpportunityRepository(db *gorm.DB) repoInterfaces.PartnershipOpportunityRepository {
	return &partnershipOpportunityRepository{db: db}
}

func (r *partnershipOpportunityRepository) Create(ctx context.Context, opp *models.PartnershipOpportunity) error {
	if err := r.db.WithContext(ctx).Create(opp).Error; err != nil {
		return fmt.Errorf("failed to create opportunity: %w", err)
	}
	return nil
}

func (r *partnershipOpportunityRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.PartnershipOpportunity, error) {
	var opp models.PartnershipOpportunity
	if err := r.db.WithContext(ctx).First(&opp, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find opportunity: %w", err)
	}
	return &opp, nil
}

func (r *partnershipOpportunityRepository) FindByEvent(ctx context.Context, eventID uuid.UUID, onlyActive bool) ([]models.PartnershipOpportunity, error) {
	var list []models.PartnershipOpportunity
	q := r.db.WithContext(ctx).Model(&models.PartnershipOpportunity{}).Where("event_id = ?", eventID)
	if onlyActive {
		q = q.Where("is_active = ?", true)
	}
	if err := q.Order("sort_order asc, created_at asc").Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to list opportunities: %w", err)
	}
	return list, nil
}

func (r *partnershipOpportunityRepository) Update(ctx context.Context, opp *models.PartnershipOpportunity) error {
	if err := r.db.WithContext(ctx).Save(opp).Error; err != nil {
		return fmt.Errorf("failed to update opportunity: %w", err)
	}
	return nil
}

func (r *partnershipOpportunityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.PartnershipOpportunity{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete opportunity: %w", err)
	}
	return nil
}

func (r *partnershipOpportunityRepository) DecrementSlots(ctx context.Context, id uuid.UUID, delta int) error {
	if delta <= 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Exec(`
		UPDATE partnership_opportunities
		SET slots_remaining = GREATEST(slots_remaining - ?, 0),
		    updated_at = NOW()
		WHERE id = ? AND slots_remaining >= ?
	`, delta, id, delta)
	if res.Error != nil {
		return fmt.Errorf("failed to decrement slots: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("no slots remaining for this opportunity")
	}
	return nil
}

func (r *partnershipOpportunityRepository) IncrementSlots(ctx context.Context, id uuid.UUID, delta int) error {
	if delta <= 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Exec(`
		UPDATE partnership_opportunities
		SET slots_remaining = LEAST(slots_remaining + ?, slots_total),
		    updated_at = NOW()
		WHERE id = ?
	`, delta, id).Error; err != nil {
		return fmt.Errorf("failed to increment slots: %w", err)
	}
	return nil
}

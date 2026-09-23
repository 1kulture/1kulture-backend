package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type brandProfileRepository struct {
	db *gorm.DB
}

func NewBrandProfileRepository(db *gorm.DB) repoInterfaces.BrandProfileRepository {
	return &brandProfileRepository{db: db}
}

func (r *brandProfileRepository) Create(ctx context.Context, profile *models.BrandProfile) error {
	if err := r.db.WithContext(ctx).Create(profile).Error; err != nil {
		return fmt.Errorf("failed to create brand profile: %w", err)
	}
	return nil
}

func (r *brandProfileRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.BrandProfile, error) {
	var p models.BrandProfile
	if err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find brand profile: %w", err)
	}
	return &p, nil
}

func (r *brandProfileRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.BrandProfile, error) {
	var p models.BrandProfile
	if err := r.db.WithContext(ctx).First(&p, "user_id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find brand profile by user: %w", err)
	}
	return &p, nil
}

func (r *brandProfileRepository) Update(ctx context.Context, profile *models.BrandProfile) error {
	if err := r.db.WithContext(ctx).Save(profile).Error; err != nil {
		return fmt.Errorf("failed to update brand profile: %w", err)
	}
	return nil
}

func (r *brandProfileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&models.BrandProfile{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete brand profile: %w", err)
	}
	return nil
}

func (r *brandProfileRepository) List(ctx context.Context, industry, location string, page, perPage int) ([]models.BrandProfile, int64, error) {
	var list []models.BrandProfile
	var total int64
	q := r.db.WithContext(ctx).Model(&models.BrandProfile{}).Where("is_active = ?", true)
	if industry != "" {
		q = q.Where("LOWER(industry) = LOWER(?)", industry)
	}
	if location != "" {
		q = q.Where("LOWER(location) LIKE LOWER(?)", "%"+location+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count brands: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Order("business_name asc").Limit(perPage).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list brands: %w", err)
	}
	return list, total, nil
}

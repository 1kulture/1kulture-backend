package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type settingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) interfaces.SettingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) Create(ctx context.Context, setting *models.Setting) error {
	if err := r.db.WithContext(ctx).Create(setting).Error; err != nil {
		return fmt.Errorf("failed to create setting: %w", err)
	}
	return nil
}

func (r *settingRepository) FindByKey(ctx context.Context, key string) (*models.Setting, error) {
	var s models.Setting
	if err := r.db.WithContext(ctx).First(&s, "key = ?", key).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find setting by key: %w", err)
	}
	return &s, nil
}

func (r *settingRepository) FindAll(ctx context.Context, category string) ([]models.Setting, error) {
	var settings []models.Setting
	q := r.db.WithContext(ctx).Model(&models.Setting{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if err := q.Order("category asc, key asc").Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("failed to list settings: %w", err)
	}
	return settings, nil
}

func (r *settingRepository) FindPublic(ctx context.Context) ([]models.Setting, error) {
	var settings []models.Setting
	if err := r.db.WithContext(ctx).Where("is_public = ?", true).Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("failed to list public settings: %w", err)
	}
	return settings, nil
}

func (r *settingRepository) Update(ctx context.Context, setting *models.Setting) error {
	if err := r.db.WithContext(ctx).Save(setting).Error; err != nil {
		return fmt.Errorf("failed to update setting: %w", err)
	}
	return nil
}

func (r *settingRepository) Upsert(ctx context.Context, setting *models.Setting) error {
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "value_type", "category", "description", "is_public", "updated_by", "updated_at"}),
		}).
		Create(setting).Error; err != nil {
		return fmt.Errorf("failed to upsert setting: %w", err)
	}
	return nil
}

func (r *settingRepository) Delete(ctx context.Context, key string) error {
	if err := r.db.WithContext(ctx).Where("key = ?", key).Delete(&models.Setting{}).Error; err != nil {
		return fmt.Errorf("failed to delete setting: %w", err)
	}
	return nil
}

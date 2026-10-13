package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

// ---------- Event Category ----------

type eventCategoryRepository struct {
	db *gorm.DB
}

func NewEventCategoryRepository(db *gorm.DB) repoInterfaces.EventCategoryRepository {
	return &eventCategoryRepository{db: db}
}

func (r *eventCategoryRepository) Create(ctx context.Context, c *models.EventCategory) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *eventCategoryRepository) FindBySlug(ctx context.Context, slug string) (*models.EventCategory, error) {
	var c models.EventCategory
	if err := r.db.WithContext(ctx).First(&c, "slug = ?", slug).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *eventCategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.EventCategory, error) {
	var c models.EventCategory
	if err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *eventCategoryRepository) FindBySlugs(ctx context.Context, slugs []string) ([]models.EventCategory, error) {
	if len(slugs) == 0 {
		return []models.EventCategory{}, nil
	}
	var list []models.EventCategory
	if err := r.db.WithContext(ctx).Where("slug IN ?", slugs).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *eventCategoryRepository) ListActive(ctx context.Context) ([]models.EventCategory, error) {
	var list []models.EventCategory
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order asc, name asc").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *eventCategoryRepository) Update(ctx context.Context, c *models.EventCategory) error {
	return r.db.WithContext(ctx).Save(c).Error
}

// ---------- User Interest ----------

type userInterestRepository struct {
	db *gorm.DB
}

func NewUserInterestRepository(db *gorm.DB) repoInterfaces.UserInterestRepository {
	return &userInterestRepository{db: db}
}

// ReplaceForUser atomically replaces all interests for a user with the given set.
func (r *userInterestRepository) ReplaceForUser(ctx context.Context, userID uuid.UUID, categoryIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&models.UserInterest{}).Error; err != nil {
			return fmt.Errorf("failed to clear interests: %w", err)
		}
		if len(categoryIDs) == 0 {
			return nil
		}
		rows := make([]models.UserInterest, 0, len(categoryIDs))
		for _, cid := range categoryIDs {
			rows = append(rows, models.UserInterest{
				UserID:     userID,
				CategoryID: cid,
			})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return fmt.Errorf("failed to insert interests: %w", err)
		}
		return nil
	})
}

func (r *userInterestRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.UserInterest, error) {
	var list []models.UserInterest
	if err := r.db.WithContext(ctx).
		Preload("Category").
		Where("user_id = ?", userID).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *userInterestRepository) ListCategoryIDsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	if err := r.db.WithContext(ctx).
		Model(&models.UserInterest{}).
		Where("user_id = ?", userID).
		Pluck("category_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *userInterestRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.UserInterest{}).
		Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
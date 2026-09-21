package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type currencyRepository struct {
	db *gorm.DB
}

func NewCurrencyRepository(db *gorm.DB) interfaces.CurrencyRepository {
	return &currencyRepository{db: db}
}

func (r *currencyRepository) Create(ctx context.Context, currency *models.Currency) error {
	if err := r.db.WithContext(ctx).Create(currency).Error; err != nil {
		return fmt.Errorf("failed to create currency: %w", err)
	}
	return nil
}

func (r *currencyRepository) FindByCode(ctx context.Context, code string) (*models.Currency, error) {
	var c models.Currency
	if err := r.db.WithContext(ctx).First(&c, "code = ?", code).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find currency by code: %w", err)
	}
	return &c, nil
}

func (r *currencyRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Currency, error) {
	var c models.Currency
	if err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find currency by id: %w", err)
	}
	return &c, nil
}

func (r *currencyRepository) FindAll(ctx context.Context, onlyEnabled bool) ([]models.Currency, error) {
	var currencies []models.Currency
	q := r.db.WithContext(ctx).Model(&models.Currency{})
	if onlyEnabled {
		q = q.Where("is_enabled = ?", true)
	}
	if err := q.Order("code asc").Find(&currencies).Error; err != nil {
		return nil, fmt.Errorf("failed to list currencies: %w", err)
	}
	return currencies, nil
}

func (r *currencyRepository) FindDefault(ctx context.Context) (*models.Currency, error) {
	var c models.Currency
	if err := r.db.WithContext(ctx).First(&c, "is_default = ?", true).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find default currency: %w", err)
	}
	return &c, nil
}

func (r *currencyRepository) Update(ctx context.Context, currency *models.Currency) error {
	if err := r.db.WithContext(ctx).Save(currency).Error; err != nil {
		return fmt.Errorf("failed to update currency: %w", err)
	}
	return nil
}

func (r *currencyRepository) SetDefault(ctx context.Context, code string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Currency{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return fmt.Errorf("failed to clear old default: %w", err)
		}
		if err := tx.Model(&models.Currency{}).Where("code = ?", code).Updates(map[string]interface{}{
			"is_default": true,
			"is_enabled": true,
		}).Error; err != nil {
			return fmt.Errorf("failed to set new default: %w", err)
		}
		return nil
	})
}

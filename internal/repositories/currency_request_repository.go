package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type currencyRequestRepository struct {
	db *gorm.DB
}

func NewCurrencyRequestRepository(db *gorm.DB) interfaces.CurrencyRequestRepository {
	return &currencyRequestRepository{db: db}
}

func (r *currencyRequestRepository) Create(ctx context.Context, req *models.CurrencyRequest) error {
	if err := r.db.WithContext(ctx).Create(req).Error; err != nil {
		return fmt.Errorf("failed to create currency request: %w", err)
	}
	return nil
}

func (r *currencyRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.CurrencyRequest, error) {
	var req models.CurrencyRequest
	if err := r.db.WithContext(ctx).Preload("Requester").First(&req, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find currency request: %w", err)
	}
	return &req, nil
}

func (r *currencyRequestRepository) FindByRequester(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.CurrencyRequest, int64, error) {
	var reqs []models.CurrencyRequest
	var total int64
	q := r.db.WithContext(ctx).Model(&models.CurrencyRequest{}).Where("requested_by = ?", userID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count currency requests: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Order("created_at desc").Limit(perPage).Offset(offset).Find(&reqs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list currency requests: %w", err)
	}
	return reqs, total, nil
}

func (r *currencyRequestRepository) FindAll(ctx context.Context, status string, page, perPage int) ([]models.CurrencyRequest, int64, error) {
	var reqs []models.CurrencyRequest
	var total int64
	q := r.db.WithContext(ctx).Model(&models.CurrencyRequest{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count currency requests: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Preload("Requester").Order("created_at desc").Limit(perPage).Offset(offset).Find(&reqs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list currency requests: %w", err)
	}
	return reqs, total, nil
}

func (r *currencyRequestRepository) FindExisting(ctx context.Context, userID uuid.UUID, code string) (*models.CurrencyRequest, error) {
	var req models.CurrencyRequest
	if err := r.db.WithContext(ctx).
		Where("requested_by = ? AND currency_code = ? AND status = ?", userID, code, models.CurrencyRequestPending).
		First(&req).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find existing currency request: %w", err)
	}
	return &req, nil
}

func (r *currencyRequestRepository) Update(ctx context.Context, req *models.CurrencyRequest) error {
	if err := r.db.WithContext(ctx).Save(req).Error; err != nil {
		return fmt.Errorf("failed to update currency request: %w", err)
	}
	return nil
}

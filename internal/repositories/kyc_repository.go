package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type kycRepository struct {
	db *gorm.DB
}

func NewKYCRepository(db *gorm.DB) interfaces.KYCRepository {
	return &kycRepository{
		db: db,
	}
}

func (r *kycRepository) Create(ctx context.Context, kyc *models.KYCVerification) error {
	if err := r.db.WithContext(ctx).Create(kyc).Error; err != nil {
		return fmt.Errorf("failed to create KYC verification: %w", err)
	}
	return nil
}

func (r *kycRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.KYCVerification, error) {
	var kyc models.KYCVerification
	if err := r.db.WithContext(ctx).First(&kyc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find KYC by id: %w", err)
	}
	return &kyc, nil
}

func (r *kycRepository) FindByUserAndRole(ctx context.Context, userID, roleID uuid.UUID) (*models.KYCVerification, error) {
	var kyc models.KYCVerification
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Order("created_at DESC").
		First(&kyc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find KYC by user and role: %w", err)
	}
	return &kyc, nil
}

func (r *kycRepository) FindLatestByUser(ctx context.Context, userID uuid.UUID) (*models.KYCVerification, error) {
	var kyc models.KYCVerification
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&kyc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find latest KYC by user: %w", err)
	}
	return &kyc, nil
}

func (r *kycRepository) Update(ctx context.Context, kyc *models.KYCVerification) error {
	if err := r.db.WithContext(ctx).Save(kyc).Error; err != nil {
		return fmt.Errorf("failed to update KYC verification: %w", err)
	}
	return nil
}

func (r *kycRepository) List(ctx context.Context, page, perPage int) ([]models.KYCVerification, int64, error) {
	var kycs []models.KYCVerification
	var total int64

	query := r.db.WithContext(ctx).Model(&models.KYCVerification{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count KYC verifications: %w", err)
	}

	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&kycs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list KYC verifications: %w", err)
	}

	return kycs, total, nil
}

func (r *kycRepository) ListByStatus(ctx context.Context, status models.KYCStatus, page, perPage int) ([]models.KYCVerification, int64, error) {
	var kycs []models.KYCVerification
	var total int64

	query := r.db.WithContext(ctx).Model(&models.KYCVerification{}).Where("status = ?", status)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count KYC verifications by status: %w", err)
	}

	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&kycs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list KYC verifications by status: %w", err)
	}

	return kycs, total, nil
}

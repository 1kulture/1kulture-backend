package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) interfaces.AuditLogRepository {
	return &auditLogRepository{
		db: db,
	}
}

func (r *auditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

func (r *auditLogRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error) {
	var auditLog models.AuditLog
	if err := r.db.WithContext(ctx).First(&auditLog, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find audit log: %w", err)
	}
	return &auditLog, nil
}

func (r *auditLogRepository) List(ctx context.Context, page, perPage int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AuditLog{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}

	return logs, total, nil
}

func (r *auditLogRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, perPage int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AuditLog{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user audit logs: %w", err)
	}

	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list user audit logs: %w", err)
	}

	return logs, total, nil
}

func (r *auditLogRepository) ListByAction(ctx context.Context, action string, page, perPage int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AuditLog{}).Where("action = ?", action)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count action audit logs: %w", err)
	}

	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list action audit logs: %w", err)
	}

	return logs, total, nil
}

package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	"github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type waitlistRepository struct {
	db *gorm.DB
}

func NewWaitlistRepository(db *gorm.DB) interfaces.WaitlistRepository {
	return &waitlistRepository{db: db}
}

func (r *waitlistRepository) Create(ctx context.Context, entry *models.WaitlistEntry) error {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return fmt.Errorf("failed to create waitlist entry: %w", err)
	}
	return nil
}

func (r *waitlistRepository) FindByEmail(ctx context.Context, email string) ([]models.WaitlistEntry, error) {
	var entries []models.WaitlistEntry
	if err := r.db.WithContext(ctx).Where("email = ?", email).Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("failed to find waitlist entries by email: %w", err)
	}
	return entries, nil
}

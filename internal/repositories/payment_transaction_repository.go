package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type paymentTransactionRepository struct {
	db *gorm.DB
}

func NewPaymentTransactionRepository(db *gorm.DB) repoInterfaces.PaymentTransactionRepository {
	return &paymentTransactionRepository{db: db}
}

func (r *paymentTransactionRepository) Create(ctx context.Context, txn *models.PaymentTransaction) error {
	if err := r.db.WithContext(ctx).Create(txn).Error; err != nil {
		return fmt.Errorf("failed to create payment transaction: %w", err)
	}
	return nil
}

func (r *paymentTransactionRepository) FindByReference(ctx context.Context, reference string) (*models.PaymentTransaction, error) {
	var txn models.PaymentTransaction
	if err := r.db.WithContext(ctx).First(&txn, "reference = ?", reference).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find payment transaction: %w", err)
	}
	return &txn, nil
}

func (r *paymentTransactionRepository) FindByOrder(ctx context.Context, orderID uuid.UUID) ([]models.PaymentTransaction, error) {
	var list []models.PaymentTransaction
	if err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at desc").
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to list payment transactions: %w", err)
	}
	return list, nil
}

func (r *paymentTransactionRepository) Update(ctx context.Context, txn *models.PaymentTransaction) error {
	if err := r.db.WithContext(ctx).Save(txn).Error; err != nil {
		return fmt.Errorf("failed to update payment transaction: %w", err)
	}
	return nil
}

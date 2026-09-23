package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/1kulture/1kulture-backend/internal/models"
	repoInterfaces "github.com/1kulture/1kulture-backend/internal/repositories/interfaces"
)

type ledgerRepository struct {
	db *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) repoInterfaces.LedgerRepository {
	return &ledgerRepository{db: db}
}

func (r *ledgerRepository) Create(ctx context.Context, entry *models.LedgerEntry) error {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return fmt.Errorf("failed to create ledger entry: %w", err)
	}
	return nil
}

func (r *ledgerRepository) CreateMany(ctx context.Context, entries []models.LedgerEntry) error {
	if len(entries) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&entries).Error; err != nil {
		return fmt.Errorf("failed to create ledger entries: %w", err)
	}
	return nil
}

// SumAccount returns total credits and debits for an account (optionally per owner).
func (r *ledgerRepository) SumAccount(ctx context.Context, account models.LedgerAccount, ownerID *uuid.UUID, currency string) (creditMinor, debitMinor int64, err error) {
	type row struct {
		EntryType string
		Total     int64
	}
	var rows []row
	q := r.db.WithContext(ctx).Model(&models.LedgerEntry{}).
		Select("entry_type, SUM(amount_minor) as total").
		Where("account = ? AND currency = ?", account, currency).
		Group("entry_type")
	if ownerID != nil {
		q = q.Where("owner_id = ?", *ownerID)
	}
	if err := q.Scan(&rows).Error; err != nil {
		return 0, 0, fmt.Errorf("failed to sum ledger account: %w", err)
	}
	for _, rw := range rows {
		switch rw.EntryType {
		case string(models.LedgerEntryCredit):
			creditMinor = rw.Total
		case string(models.LedgerEntryDebit):
			debitMinor = rw.Total
		}
	}
	return creditMinor, debitMinor, nil
}

func (r *ledgerRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, page, perPage int) ([]models.LedgerEntry, int64, error) {
	var list []models.LedgerEntry
	var total int64
	q := r.db.WithContext(ctx).Model(&models.LedgerEntry{}).Where("owner_id = ?", ownerID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count ledger entries: %w", err)
	}
	offset := (page - 1) * perPage
	if err := q.Order("created_at desc").Limit(perPage).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list ledger entries: %w", err)
	}
	return list, total, nil
}

func (r *ledgerRepository) ListByReference(ctx context.Context, reference string) ([]models.LedgerEntry, error) {
	var list []models.LedgerEntry
	if err := r.db.WithContext(ctx).
		Where("reference = ?", reference).
		Order("created_at asc").
		Find(&list).Error; err != nil {
		return nil, fmt.Errorf("failed to list ledger entries by reference: %w", err)
	}
	return list, nil
}

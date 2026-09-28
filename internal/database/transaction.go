package database

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

// WithTx stores a transaction on the context so repositories can detect it.
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// FromContext returns the transaction from the context, or nil.
func FromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return nil
}

// InTransaction wraps fn in a DB transaction and makes the tx visible
// to any repository method that uses database.FromContext(ctx).
func InTransaction(ctx context.Context, db *gorm.DB, fn func(txCtx context.Context) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(WithTx(ctx, tx))
	})
}

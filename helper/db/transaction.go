package db

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

func (d *DB) Get(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return d.conn.WithContext(ctx)
}

// WithTransaction runs fn inside a transaction.
func (d *DB) WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	tx := d.conn.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

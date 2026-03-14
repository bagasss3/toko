package repository

import (
	"context"

	"github.com/bagasss3/toko/packages/database"

	"github.com/bagasss3/toko/services/auth-service/internal/model"
	"gorm.io/gorm"
)

type (
	gormTransactioner struct {
		db *database.DB
	}
)

// NewGormTransactioner
func NewGormTransactioner(db *database.DB) model.GormTransactioner {
	return &gormTransactioner{db: db}
}

// Begin
func (t *gormTransactioner) Begin(ctx context.Context) *gorm.DB {
	return t.db.Conn.WithContext(ctx).Begin()
}

// Commit
func (t *gormTransactioner) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}

// Rollback
func (t *gormTransactioner) Rollback(tx *gorm.DB) {
	tx.Rollback()
}

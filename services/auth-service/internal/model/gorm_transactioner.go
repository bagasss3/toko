package model

import (
	"context"

	"gorm.io/gorm"
)

type GormTransactioner interface {
	Begin(ctx context.Context) *gorm.DB
	Commit(tx *gorm.DB) error
	Rollback(tx *gorm.DB)
}

package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/bagasss3/toko/packages/database"
	"github.com/bagasss3/toko/services/auth-service/internal/model"
)

type addressRepository struct {
	db *database.DB
}

func NewAddressRepository(db *database.DB) model.AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Address, error) {
	var address model.Address
	err := r.db.Conn.WithContext(ctx).First(&address, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &address, err
}

func (r *addressRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Address, error) {
	var addresses []*model.Address
	err := r.db.Conn.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&addresses).Error
	return addresses, err
}

func (r *addressRepository) Create(ctx context.Context, address *model.Address) error {
	return r.db.Conn.WithContext(ctx).Create(address).Error
}

func (r *addressRepository) CreateWithTx(ctx context.Context, tx *gorm.DB, address *model.Address) error {
	return tx.WithContext(ctx).Create(address).Error
}

func (r *addressRepository) Update(ctx context.Context, address *model.Address) error {
	return r.db.Conn.WithContext(ctx).Save(address).Error
}

func (r *addressRepository) UpdateWithTx(ctx context.Context, tx *gorm.DB, address *model.Address) error {
	return tx.WithContext(ctx).Save(address).Error
}

func (r *addressRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.Conn.WithContext(ctx).Delete(&model.Address{}, "id = ?", id).Error
}

func (r *addressRepository) DeleteWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) error {
	return tx.WithContext(ctx).Delete(&model.Address{}, "id = ?", id).Error
}

func (r *addressRepository) SetDefault(ctx context.Context, userID, addressID uuid.UUID) error {
	return r.db.Conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unset current default
		if err := tx.Model(&model.Address{}).
			Where("user_id = ?", userID).
			Update("is_default", false).Error; err != nil {
			return err
		}
		// Set new default
		return tx.Model(&model.Address{}).
			Where("id = ? AND user_id = ?", addressID, userID).
			Update("is_default", true).Error
	})
}

package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	apperrors "github.com/bagasss3/toko/packages/errors"
	"github.com/bagasss3/toko/services/auth-service/internal/model"
)

type addressUsecase struct {
	addressRepo   model.AddressRepository
	transactioner model.GormTransactioner
}

func NewAddressUsecase(
	addressRepo model.AddressRepository,
	transactioner model.GormTransactioner,
) model.AddressUsecase {
	return &addressUsecase{
		addressRepo:   addressRepo,
		transactioner: transactioner,
	}
}

func (u *addressUsecase) GetMyAddresses(ctx context.Context, userID uuid.UUID) ([]*model.Address, error) {
	return u.addressRepo.FindByUserID(ctx, userID)
}

func (u *addressUsecase) GetByID(ctx context.Context, userID, addressID uuid.UUID) (*model.Address, error) {
	address, err := u.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		return nil, fmt.Errorf("finding address: %w", err)
	}
	if address == nil {
		return nil, apperrors.ErrAddressNotFound
	}
	if address.UserID != userID {
		return nil, apperrors.ErrNotOwner
	}
	return address, nil
}

func (u *addressUsecase) Create(ctx context.Context, userID uuid.UUID, req model.AddressRequest) (*model.Address, error) {
	address := &model.Address{
		UserID:       userID,
		ReceiverName: req.ReceiverName,
		Phone:        req.Phone,
		AddressLine:  req.AddressLine,
		City:         req.City,
		Province:     req.Province,
		PostalCode:   req.PostalCode,
		IsDefault:    req.IsDefault,
	}

	// If this is the first address or is_default is true, handle default logic in transaction
	if req.IsDefault {
		tx := u.transactioner.Begin(ctx)
		defer func() {
			if r := recover(); r != nil {
				u.transactioner.Rollback(tx)
				panic(r)
			}
		}()

		// Unset other defaults
		addresses, err := u.addressRepo.FindByUserID(ctx, userID)
		if err != nil {
			u.transactioner.Rollback(tx)
			return nil, fmt.Errorf("finding addresses: %w", err)
		}

		if len(addresses) > 0 {
			for _, addr := range addresses {
				if addr.IsDefault {
					addr.IsDefault = false
					if err := u.addressRepo.UpdateWithTx(ctx, tx, addr); err != nil {
						u.transactioner.Rollback(tx)
						return nil, fmt.Errorf("updating address: %w", err)
					}
				}
			}
		}

		if err := u.addressRepo.CreateWithTx(ctx, tx, address); err != nil {
			u.transactioner.Rollback(tx)
			return nil, fmt.Errorf("creating address: %w", err)
		}

		if err := u.transactioner.Commit(tx); err != nil {
			return nil, fmt.Errorf("committing transaction: %w", err)
		}
	} else {
		if err := u.addressRepo.Create(ctx, address); err != nil {
			return nil, fmt.Errorf("creating address: %w", err)
		}
	}

	return address, nil
}

func (u *addressUsecase) Update(ctx context.Context, userID, addressID uuid.UUID, req model.AddressRequest) (*model.Address, error) {
	address, err := u.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		return nil, fmt.Errorf("finding address: %w", err)
	}
	if address == nil {
		return nil, apperrors.ErrAddressNotFound
	}
	if address.UserID != userID {
		return nil, apperrors.ErrNotOwner
	}

	address.ReceiverName = req.ReceiverName
	address.Phone = req.Phone
	address.AddressLine = req.AddressLine
	address.City = req.City
	address.Province = req.Province
	address.PostalCode = req.PostalCode

	// Handle default change in transaction
	if req.IsDefault && !address.IsDefault {
		tx := u.transactioner.Begin(ctx)
		defer func() {
			if r := recover(); r != nil {
				u.transactioner.Rollback(tx)
				panic(r)
			}
		}()

		// Unset other defaults
		addresses, err := u.addressRepo.FindByUserID(ctx, userID)
		if err != nil {
			u.transactioner.Rollback(tx)
			return nil, fmt.Errorf("finding addresses: %w", err)
		}

		for _, addr := range addresses {
			if addr.ID != addressID && addr.IsDefault {
				addr.IsDefault = false
				if err := u.addressRepo.UpdateWithTx(ctx, tx, addr); err != nil {
					u.transactioner.Rollback(tx)
					return nil, fmt.Errorf("updating address: %w", err)
				}
			}
		}

		address.IsDefault = true
		if err := u.addressRepo.UpdateWithTx(ctx, tx, address); err != nil {
			u.transactioner.Rollback(tx)
			return nil, fmt.Errorf("updating address: %w", err)
		}

		if err := u.transactioner.Commit(tx); err != nil {
			return nil, fmt.Errorf("committing transaction: %w", err)
		}
	} else {
		address.IsDefault = req.IsDefault
		if err := u.addressRepo.Update(ctx, address); err != nil {
			return nil, fmt.Errorf("updating address: %w", err)
		}
	}

	return address, nil
}

func (u *addressUsecase) Delete(ctx context.Context, userID, addressID uuid.UUID) error {
	address, err := u.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		return fmt.Errorf("finding address: %w", err)
	}
	if address == nil {
		return apperrors.ErrAddressNotFound
	}
	if address.UserID != userID {
		return apperrors.ErrNotOwner
	}

	return u.addressRepo.Delete(ctx, addressID)
}

func (u *addressUsecase) SetDefault(ctx context.Context, userID, addressID uuid.UUID) error {
	address, err := u.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		return fmt.Errorf("finding address: %w", err)
	}
	if address == nil {
		return apperrors.ErrAddressNotFound
	}
	if address.UserID != userID {
		return apperrors.ErrNotOwner
	}

	return u.addressRepo.SetDefault(ctx, userID, addressID)
}

package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Address struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	UserID       uuid.UUID `json:"user_id"`
	ReceiverName string    `json:"receiver_name,omitempty"`
	Phone        string    `json:"phone,omitempty"`
	AddressLine  string    `json:"address_line,omitempty"`
	City         string    `json:"city,omitempty"`
	Province     string    `json:"province,omitempty"`
	PostalCode   string    `json:"postal_code,omitempty"`
	IsDefault    bool      `json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
}

type AddressRequest struct {
	ReceiverName string
	Phone        string
	AddressLine  string
	City         string
	Province     string
	PostalCode   string
	IsDefault    bool
}

type AddressRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Address, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*Address, error)
	Create(ctx context.Context, address *Address) error
	CreateWithTx(ctx context.Context, tx *gorm.DB, address *Address) error
	Update(ctx context.Context, address *Address) error
	UpdateWithTx(ctx context.Context, tx *gorm.DB, address *Address) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteWithTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) error
	SetDefault(ctx context.Context, userID, addressID uuid.UUID) error
}

type AddressUsecase interface {
	GetMyAddresses(ctx context.Context, userID uuid.UUID) ([]*Address, error)
	GetByID(ctx context.Context, userID, addressID uuid.UUID) (*Address, error)
	Create(ctx context.Context, userID uuid.UUID, req AddressRequest) (*Address, error)
	Update(ctx context.Context, userID, addressID uuid.UUID, req AddressRequest) (*Address, error)
	Delete(ctx context.Context, userID, addressID uuid.UUID) error
	SetDefault(ctx context.Context, userID, addressID uuid.UUID) error
}

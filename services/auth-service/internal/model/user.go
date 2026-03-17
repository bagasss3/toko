package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleCustomer Role = "customer"
)

type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Password hash, never expose in JSON
	Role      Role      `json:"role"`
	Phone     string    `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	Create(ctx context.Context, user *User) error
	CreateWithTx(ctx context.Context, tx *gorm.DB, user *User) error
	CountAdmins(ctx context.Context) (int64, error)
}

type UserUsecase interface {
	Register(ctx context.Context, req RegisterRequest) (*User, error)
	CreateAdmin(ctx context.Context, req RegisterRequest, createdBy uuid.UUID) (*User, error)
	CreateFirstAdmin(ctx context.Context) error
}

type RegisterRequest struct {
	Name     string
	Email    string
	Password string
	Phone    string
	Role     Role
	Address  *AddressRequest
}

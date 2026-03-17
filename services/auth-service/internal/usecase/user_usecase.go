package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/bagasss3/toko/packages/errors"
	"github.com/bagasss3/toko/services/auth-service/internal/model"
)

type userUsecase struct {
	userRepo      model.UserRepository
	addressRepo   model.AddressRepository
	transactioner model.GormTransactioner
}

func NewUserUsecase(
	userRepo model.UserRepository,
	addressRepo model.AddressRepository,
	transactioner model.GormTransactioner,
) model.UserUsecase {
	return &userUsecase{
		userRepo:      userRepo,
		addressRepo:   addressRepo,
		transactioner: transactioner,
	}
}

func (u *userUsecase) Register(ctx context.Context, req model.RegisterRequest) (*model.User, error) {
	if req.Role != "" && req.Role != model.RoleCustomer {
		return nil, errors.New("invalid role for customer registration")
	}
	req.Role = model.RoleCustomer

	return u.createUserWithAddress(ctx, req)
}

func (u *userUsecase) CreateAdmin(ctx context.Context, req model.RegisterRequest, createdBy uuid.UUID) (*model.User, error) {
	creator, err := u.userRepo.FindByID(ctx, createdBy)
	if err != nil {
		return nil, fmt.Errorf("finding creator: %w", err)
	}
	if creator == nil || creator.Role != model.RoleAdmin {
		return nil, apperrors.ErrUnauthorized
	}

	if req.Role != model.RoleAdmin {
		return nil, errors.New("role must be admin")
	}

	return u.createUserWithAddress(ctx, req)
}

func (u *userUsecase) createUserWithAddress(ctx context.Context, req model.RegisterRequest) (*model.User, error) {
	existing, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("checking existing user: %w", err)
	}
	if existing != nil {
		return nil, apperrors.ErrEmailExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &model.User{
		ID:       uuid.New(),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashed),
		Role:     req.Role,
		Phone:    req.Phone,
	}

	if req.Address == nil {
		if err := u.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("creating user: %w", err)
		}
		return user, nil
	}

	tx := u.transactioner.Begin(ctx)
	if err := u.userRepo.CreateWithTx(ctx, tx, user); err != nil {
		u.transactioner.Rollback(tx)
		return nil, fmt.Errorf("creating user: %w", err)
	}

	address := &model.Address{
		ID:           uuid.New(),
		UserID:       user.ID,
		ReceiverName: req.Address.ReceiverName,
		Phone:        req.Address.Phone,
		AddressLine:  req.Address.AddressLine,
		City:         req.Address.City,
		Province:     req.Address.Province,
		PostalCode:   req.Address.PostalCode,
		IsDefault:    req.Address.IsDefault,
	}
	if err := u.addressRepo.CreateWithTx(ctx, tx, address); err != nil {
		u.transactioner.Rollback(tx)
		return nil, fmt.Errorf("creating address: %w", err)
	}

	if err := u.transactioner.Commit(tx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return user, nil
}

func (u *userUsecase) CreateFirstAdmin(ctx context.Context) error {
	count, err := u.userRepo.CountAdmins(ctx)
	if err != nil {
		return fmt.Errorf("checking admin count: %w", err)
	}

	if count > 0 {
		log.Info("admin account already exists, skipping first admin creation")
		return nil
	}

	req := model.RegisterRequest{
		Name:     "Admin",
		Email:    "admin@toko.com",
		Password: "admin123",
		Role:     model.RoleAdmin,
		Phone:    "",
	}

	admin, err := u.createUserWithAddress(ctx, req)
	if err != nil {
		return fmt.Errorf("creating first admin: %w", err)
	}

	log.WithFields(log.Fields{
		"user_id": admin.ID,
		"email":   admin.Email,
	}).Warn("First admin account created. Please change the default password immediately!")

	return nil
}

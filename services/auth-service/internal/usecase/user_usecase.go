package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

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
	// Customers can only register as customers
	if req.Role != "" && req.Role != model.RoleCustomer {
		return nil, errors.New("invalid role for customer registration")
	}
	req.Role = model.RoleCustomer

	return u.createUserWithAddress(ctx, req)
}

func (u *userUsecase) CreateAdmin(ctx context.Context, req model.RegisterRequest, createdBy uuid.UUID) (*model.User, error) {
	// Verify creator is an admin
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
	// Start transaction
	tx := u.transactioner.Begin(ctx)
	defer func() {
		if r := recover(); r != nil {
			u.transactioner.Rollback(tx)
			panic(r)
		}
	}()

	// Check existing email using transaction
	var existing *model.User
	var err error
	existing, err = u.findByEmailTx(ctx, tx, req.Email)
	if err != nil {
		u.transactioner.Rollback(tx)
		return nil, fmt.Errorf("checking existing user: %w", err)
	}
	if existing != nil {
		u.transactioner.Rollback(tx)
		return nil, apperrors.ErrEmailExists
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		u.transactioner.Rollback(tx)
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	// Create user
	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashed),
		Role:     req.Role,
		Phone:    req.Phone,
	}

	if err := u.userRepo.CreateWithTx(ctx, tx, user); err != nil {
		u.transactioner.Rollback(tx)
		return nil, fmt.Errorf("creating user: %w", err)
	}

	// Create address if provided
	if req.Address != nil {
		address := &model.Address{
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
	}

	// Commit transaction
	if err := u.transactioner.Commit(tx); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return user, nil
}

// Helper to find user by email within transaction
func (u *userUsecase) findByEmailTx(ctx context.Context, tx *gorm.DB, email string) (*model.User, error) {
	var user model.User
	err := tx.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

// CreateFirstAdmin creates the first admin account if no admins exist
func (u *userUsecase) CreateFirstAdmin(ctx context.Context) error {
	// Check if any admin exists
	count, err := u.userRepo.CountAdmins(ctx)
	if err != nil {
		return fmt.Errorf("checking admin count: %w", err)
	}

	if count > 0 {
		log.Info("admin account already exists, skipping first admin creation")
		return nil
	}

	// Create default admin
	req := model.RegisterRequest{
		Name:     "Admin",
		Email:    "admin@toko.com",
		Password: "admin123", // Should be changed immediately after first login
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

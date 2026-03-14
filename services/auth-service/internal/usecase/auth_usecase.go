package usecase

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/bagasss3/toko/services/auth-service/internal/config"
	"github.com/bagasss3/toko/services/auth-service/internal/helper/token"
	"github.com/bagasss3/toko/services/auth-service/internal/model"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type authUsecase struct {
	userRepo   model.UserRepository
	tokenMaker *token.Maker
	cfg        config.Config
}

func NewAuthUsecase(
	userRepo model.UserRepository,
	tokenMaker *token.Maker,
	cfg config.Config,
) model.AuthUsecase {
	return &authUsecase{
		userRepo:   userRepo,
		tokenMaker: tokenMaker,
		cfg:        cfg,
	}
}

func (u *authUsecase) Login(ctx context.Context, email, password string) (*model.TokenPair, error) {
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, _, err := u.tokenMaker.CreateToken(user.ID, user.Email, u.cfg.AccessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("creating token: %w", err)
	}

	log.WithField("user_id", user.ID).Info("user logged in")

	return &model.TokenPair{
		AccessToken: accessToken,
	}, nil
}

func (u *authUsecase) Register(ctx context.Context, email, password string) (*model.User, error) {
	existing, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("checking existing user: %w", err)
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := &model.User{
		Email:    email,
		Password: string(hashed),
	}
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	return user, nil
}

func (u *authUsecase) ValidateToken(ctx context.Context, tokenStr string) (*model.TokenClaims, error) {
	claims, err := u.tokenMaker.VerifyToken(tokenStr)
	if err != nil {
		return nil, err
	}

	return &model.TokenClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
	}, nil
}

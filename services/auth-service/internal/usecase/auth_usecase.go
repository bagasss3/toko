package usecase

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	apperrors "github.com/bagasss3/toko/packages/errors"
	"github.com/bagasss3/toko/services/auth-service/internal/config"
	"github.com/bagasss3/toko/services/auth-service/internal/helper/token"
	"github.com/bagasss3/toko/services/auth-service/internal/model"
)

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
		return nil, apperrors.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	accessToken, _, err := u.tokenMaker.CreateToken(user.ID, user.Email, user.Role, u.cfg.AccessTokenDuration)
	if err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken: accessToken,
	}, nil
}

func (u *authUsecase) ValidateToken(ctx context.Context, tokenStr string) (*model.TokenClaims, error) {
	claims, err := u.tokenMaker.VerifyToken(tokenStr)
	if err != nil {
		return nil, err
	}

	return &model.TokenClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}

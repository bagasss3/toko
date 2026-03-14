package model

import (
	"context"

	"github.com/google/uuid"
)

type AuthUsecase interface {
	Login(ctx context.Context, email, password string) (*TokenPair, error)
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   Role      `json:"role"`
}

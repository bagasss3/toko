package model

import "context"

type AuthUsecase interface {
    Login(ctx context.Context, email, password string) (*TokenPair, error)
    Register(ctx context.Context, email, password string) (*User, error)
    ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
}

type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
}

type TokenClaims struct {
    UserID int64  `json:"user_id"`
    Email  string `json:"email"`
}
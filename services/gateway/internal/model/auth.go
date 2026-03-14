package model

type TokenClaims struct {
	UserID int64
	Email  string
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

type UserResponse struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
}

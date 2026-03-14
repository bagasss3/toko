package model

type TokenClaims struct {
	UserID string
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
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

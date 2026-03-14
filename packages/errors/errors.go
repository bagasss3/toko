package errors

import "errors"

// Auth errors
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token has expired")
)

// User errors
var (
	ErrEmailExists     = errors.New("email already registered")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidRole     = errors.New("invalid role")
)

// Address errors
var (
	ErrAddressNotFound = errors.New("address not found")
	ErrNotOwner        = errors.New("you don't own this address")
)

// Common errors
var (
	ErrNotFound        = errors.New("not found")
	ErrAlreadyExists   = errors.New("already exists")
	ErrValidation      = errors.New("validation error")
	ErrInternal        = errors.New("internal error")
)

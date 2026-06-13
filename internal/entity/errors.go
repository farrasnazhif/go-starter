package entity

import "errors"

var (
	ErrNotFound            = errors.New("resource not found")
	ErrDuplicateEmail      = errors.New("a user with that email already exists")
	ErrDuplicateUsername   = errors.New("a user with that username already exists")
	ErrOTPExpired          = errors.New("otp expired")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrAccountInactive     = errors.New("account not activated")
	ErrEmailExists         = errors.New("email already registered")
	ErrInsufficientCredits = errors.New("insufficient credits")
)

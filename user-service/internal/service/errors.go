package service

import "errors"

var (
	ErrUserExists          = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCode         = errors.New("invalid or expired verification code")
	ErrPasswordNotSet      = errors.New("password not set")
	ErrIncorrectPassword   = errors.New("incorrect password")
	ErrAddressNotFound     = errors.New("address not found")
	ErrAddressLimitReached = errors.New("address limit reached (max 20)")
)

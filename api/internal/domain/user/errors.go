package user

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidInput       = errors.New("invalid input")
	ErrTermsNotAccepted   = errors.New("terms not accepted")
)

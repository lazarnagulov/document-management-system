package auth

import "errors"

var (
	ErrNotFound           = errors.New("user not found")
	ErrEmailTaken         = errors.New("email already in use")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

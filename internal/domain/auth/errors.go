package auth

import "errors"

var (
	ErrIDRequired = errors.New("ErrIDRequired")
	ErrEmailRequired = errors.New("ErrEmailRequired")
	ErrPasswordRequired = errors.New("ErrPasswordRequired")
	ErrRoleInvalid = errors.New("ErrRoleInvalid")
	ErrStatusInvalid = errors.New("ErrStatusInvalid")

	ErrUserNotFound = errors.New("User not found")
)
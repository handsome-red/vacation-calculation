package user

import "errors"

var (
	ErrEmailRequired      = errors.New("email is required")
	ErrInvalidEmailFormat = errors.New("invalid email format")
	ErrEmailAlreadyExists = errors.New("email already exists")

	ErrPasswordRequired  = errors.New("password is required")
	ErrPasswordTooShort  = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong   = errors.New("password must be at most 72 characters")
	ErrPasswordNoUpper   = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLower   = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoNumber  = errors.New("password must contain at least one number")
	ErrPasswordNoSpecial = errors.New("password must contain at least one special character")
	ErrPasswordHashing   = errors.New("failed to hash password")

	ErrNameRequired = errors.New("name is required")

	ErrBirthDateInvalid = errors.New("birth date inlavid")

	ErrPositionInvalid = errors.New("position invalid")

	ErrHiredDateInvalid = errors.New("hired date invalid")

	ErrIDEmpty    = errors.New("user ID cannot be empty")
	ErrIDRequired = errors.New("ID is required")

	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyActive  = errors.New("user already active")
	ErrUserAlreadyDeleted = errors.New("user already deleted")
)

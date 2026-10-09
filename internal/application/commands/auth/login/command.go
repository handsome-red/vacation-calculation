package login

import "github.com/handsome-red/vacation-calculation/internal/domain/auth"

type Command struct {
	Password string
	Email    auth.Email
}

type Result struct {
	UserID string
	Email  string
}

package get_current_user

import "github.com/handsome-red/vacation-calculation/internal/domain/auth"

type Query struct {
	SessionID string
}

type Result struct {
	User auth.AuthUser
}

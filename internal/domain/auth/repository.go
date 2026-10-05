package auth

import "context"

type AuthRepository interface {
	FindByEmail(ctx context.Context, email Email)(*AuthUser, error)
	Save(ctx context.Context, u *AuthUser) error
}
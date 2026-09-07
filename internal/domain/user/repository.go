package user

import "context"

type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, userID UserID) (*User, error)
	FindByEmail(ctx context.Context, email Email) (*User, error)
	FindActive(ctx context.Context) ([]*User, error)
	Delete(ctx context.Context, userID UserID) error
}

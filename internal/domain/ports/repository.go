package ports

import (
	"context"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type UserCommandRepository interface {
	Save(ctx context.Context, user *user.User) error
	// Update(ctx context.Context, user *user.User) error
	Delete(ctx context.Context, id user.UserID) error
}

type UserQueryRepository interface {
	FindByID(ctx context.Context, id user.UserID) (*user.User, error)
	FindByEmail(ctx context.Context, email user.Email) (*user.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	FindAll(ctx context.Context) ([]*user.User, error)
}

type UserRepository interface {
	UserCommandRepository
	UserQueryRepository
}

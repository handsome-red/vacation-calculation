package di

import (
	"context"
	"database/sql"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/activate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/deactivate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/get_active_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/get_user"
	"github.com/handsome-red/vacation-calculation/internal/config"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

type Container struct {
	// Infrastructure
	DB     *sql.DB
	Logger ports.Logger
	Hasher ports.PasswordHasher

	// Repositories
	UserRepo ports.UserRepository
	// VacationRepo ports.VacationRepository

	// Use Cases - Commands (User)
	RegisterUserUseCase   *register_user.Handler
	DeactivateUserUseCase *deactivate_user.Handler
	ActivateUserUseCase   *activate_user.Handler

	// Use Cases - Commands (Vacation)
	// CreateVacationUseCase  *create_vacation.Handler
	// ApproveVacationUseCase *approve_vacation.Handler

	// Use Cases - Queries (User)
	GetUserUseCase        *get_user.Handler
	GetActiveUsersUseCase *get_active_users.Handler
}

func NewContainer(ctx context.Context, cfg config.Config, log ports.Logger) (*Container, error) {
	return &Container{}, nil
}

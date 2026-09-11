package di

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/activate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/deactivate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/get_active_users"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/get_user"
	"github.com/handsome-red/vacation-calculation/internal/config"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/hasher"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/persistence/sqlite"
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

	db, err := sqlite.NewDB(ctx, sqlite.Config{Path: cfg.DatabasePath}, log)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	hasher := hasher.NewBcryptHasher(10)

	var userRepo ports.UserRepository = sqlite.NewUserRepository(db)

	registerUserUseCase := register_user.NewHandler(userRepo, hasher, log)
	deactivateUserUseCase := deactivate_user.NewHandler(userRepo, log)
	activateUserUseCase := activate_user.NewHandler(userRepo, log)

	getUserUseCase := get_user.NewHandler(userRepo, log)
	getActiveUsersUseCase := get_active_users.NewHandler(userRepo, log)

	return &Container{
		// Infrastructure
		DB:     db,
		Logger: log,
		Hasher: hasher,

		// Repositories
		UserRepo: userRepo,

		// Use Cases - Commands
		RegisterUserUseCase:   registerUserUseCase,
		DeactivateUserUseCase: deactivateUserUseCase,
		ActivateUserUseCase:   activateUserUseCase,

		// Use Cases - Queries
		GetUserUseCase:        getUserUseCase,
		GetActiveUsersUseCase: getActiveUsersUseCase,
	}, nil
}

func (c *Container) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

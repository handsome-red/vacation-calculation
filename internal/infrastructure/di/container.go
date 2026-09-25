package di

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/activate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/deactivate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/create_vacation"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/get_vacation_form"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/get_active_users"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/get_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/register_form"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_user_vacations"

	"github.com/handsome-red/vacation-calculation/internal/config"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/hasher"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/persistence/sqlite"
	"github.com/jmoiron/sqlx"
)

type Container struct {
	// Infrastructure
	DB     *sqlx.DB
	Logger ports.Logger
	Hasher ports.PasswordHasher

	// Repositories
	UserRepo       ports.UserRepository
	DistrictRepo   user.DistrictRepository
	DepartmentRepo user.DepartmentRepository
	// VacationRepo ports.VacationRepository

	// Use Cases - Commands (User)
	RegisterUserUseCase    *register_user.Handler
	DeactivateUserUseCase  *deactivate_user.Handler
	ActivateUserUseCase    *activate_user.Handler
	GetRegisterFormUseCase *register_form.Handler
	// Use Cases - Commands (Vacation)
	// CreateVacationUseCase  *create_vacation.Handler
	// ApproveVacationUseCase *approve_vacation.Handler

	CreateVacationUseCase *create_vacation.Handler
	GetUserVacationsUseCase  *get_user_vacations.Handler
	GetVacationFormUseCase *get_vacation_form.Handler

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
	var districtRepo user.DistrictRepository = sqlite.NewDistrictRepository(db)
	var departmentRepo user.DepartmentRepository = sqlite.NewDepartmentRepository(db)

	var vacationRepo vacation.VacationRepository = sqlite.NewVacationRepository(db)

	registerUserUseCase := register_user.NewHandler(userRepo, hasher, log)
	deactivateUserUseCase := deactivate_user.NewHandler(userRepo, log)
	activateUserUseCase := activate_user.NewHandler(userRepo, log)
	getRegisterFormUseCase := register_form.NewHandler(districtRepo, departmentRepo)

	getUserUseCase := get_user.NewHandler(userRepo, log)
	getActiveUsersUseCase := get_active_users.NewHandler(userRepo, log)

	createVacationUseCase := create_vacation.NewHandler(vacationRepo)
	getUserVacationsUseCase := get_user_vacations.NewHandler(vacationRepo)
	getVacationFormUseCase := get_vacation_form.NewHandler()

	return &Container{
		// Infrastructure
		DB:     db,
		Logger: log,
		Hasher: hasher,
		// Repositories
		UserRepo: userRepo,
		// Use Cases - Commands
		RegisterUserUseCase:    registerUserUseCase,
		DeactivateUserUseCase:  deactivateUserUseCase,
		ActivateUserUseCase:    activateUserUseCase,
		GetRegisterFormUseCase: getRegisterFormUseCase,
		// Use Cases - Queries
		GetUserUseCase:          getUserUseCase,
		GetActiveUsersUseCase:   getActiveUsersUseCase,
		DistrictRepo:            districtRepo,
		DepartmentRepo:          departmentRepo,
		CreateVacationUseCase:   createVacationUseCase,
		GetUserVacationsUseCase: getUserVacationsUseCase,
		GetVacationFormUseCase: getVacationFormUseCase,
	}, nil
}

func (c *Container) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

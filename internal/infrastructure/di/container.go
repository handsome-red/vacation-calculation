package di

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/shift/create_shift"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/activate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/deactivate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/create_vacation"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/new_holiday"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/shift/shift_form"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/get_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/list_users"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/register_form"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_calendar"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_user_vacations"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_vacation_form"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/new_holiday_form"

	"github.com/handsome-red/vacation-calculation/internal/config"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
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
	DistrictRepo   ports.DistrictRepository
	DepartmentRepo ports.DepartmentRepository
	// VacationRepo ports.VacationRepository

	// Use Cases - Commands (User)
	RegisterUserUseCase    *register_user.Handler
	DeactivateUserUseCase  *deactivate_user.Handler
	ActivateUserUseCase    *activate_user.Handler
	GetRegisterFormUseCase *register_form.Handler
	// Use Cases - Commands (Vacation)
	// ApproveVacationUseCase *approve_vacation.Handler

	CreateVacationUseCase   *create_vacation.Handler
	GetUserVacationsUseCase *get_user_vacations.Handler
	GetVacationFormUseCase  *get_vacation_form.Handler

	// Use Cases - Queries (User)
	GetUserUseCase        *get_user.Handler
	ListUsersUseCase      *list_users.Handler
	GetCalendarUseCase    *get_calendar.Handler
	NewHolidayFormUseCase *new_holiday_form.Handler
	NewHolidayUseCase     *new_holiday.Handler

	// Use Cases - Commands (Shift)
	CreateShiftUseCase *create_shift.Handler

	// Use Cases - Queries (Shift)
	ShiftFormUseCase *shift_form.Handler
}

func NewContainer(ctx context.Context, cfg config.Config, log ports.Logger) (*Container, error) {

	db, err := sqlite.NewDB(ctx, sqlite.Config{Path: cfg.DatabasePath}, log)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	hasher := hasher.NewBcryptHasher(10)

	wc := vacation.NewWorkYearCalculator()

	var userRepo ports.UserRepository = sqlite.NewUserRepository(db)
	var districtRepo ports.DistrictRepository = sqlite.NewDistrictRepository(db)
	var departmentRepo ports.DepartmentRepository = sqlite.NewDepartmentRepository(db)
	var shiftRepo ports.ShiftRepository = sqlite.NewShiftRepository(db)

	var vacationRepo ports.VacationRepository = sqlite.NewVacationRepository(db)
	var holidayRepo ports.HolidayRepository = sqlite.NewHolidayRepository(db)

	registerUserUseCase := register_user.NewHandler(userRepo, hasher, log)
	deactivateUserUseCase := deactivate_user.NewHandler(userRepo, log)
	activateUserUseCase := activate_user.NewHandler(userRepo, log)
	getRegisterFormUseCase := register_form.NewHandler(districtRepo, departmentRepo)

	getUserUseCase := get_user.NewHandler(userRepo, vacationRepo, shiftRepo, wc, log)
	getActiveUsersUseCase := list_users.NewHandler(userRepo, log)

	createVacationUseCase := create_vacation.NewHandler(vacationRepo)
	getUserVacationsUseCase := get_user_vacations.NewHandler(vacationRepo)
	getVacationFormUseCase := get_vacation_form.NewHandler()
	getCalendarUseCase := get_calendar.NewHandler(holidayRepo, vacationRepo, shiftRepo)
	newHolidayFormUseCase := new_holiday_form.NewHandler(holidayRepo)

	newHolidayUseCase := new_holiday.NewHandler(holidayRepo)

	createShiftUseCase := create_shift.NewHandler(shiftRepo)
	shiftFormUseCase := shift_form.NewHandler()

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
		ListUsersUseCase:        getActiveUsersUseCase,
		DistrictRepo:            districtRepo,
		DepartmentRepo:          departmentRepo,
		CreateVacationUseCase:   createVacationUseCase,
		GetUserVacationsUseCase: getUserVacationsUseCase,
		GetVacationFormUseCase:  getVacationFormUseCase,
		GetCalendarUseCase:      getCalendarUseCase,
		NewHolidayFormUseCase:   newHolidayFormUseCase,
		NewHolidayUseCase:       newHolidayUseCase,
		CreateShiftUseCase:      createShiftUseCase,
		ShiftFormUseCase:        shiftFormUseCase,
	}, nil
}

func (c *Container) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

package di

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/auth/login"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/calendar/new_exception"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/shift/create_shift"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/activate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/bootstrap_admin"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/deactivate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/approve_vacation"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/change_status"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/create_vacations"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/send_vacation_reminders"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/transfer_vacation"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/shift/shift_form"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/get_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/list_users"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/register_form"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/export_report"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_calendar"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_transfer_form"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_user_calendar"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_user_vacations"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_vacation_form"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/new_calendar_day_form"

	"github.com/handsome-red/vacation-calculation/internal/config"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/hasher"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/mailer"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/persistence/sqlite"
	"github.com/jmoiron/sqlx"

	"github.com/handsome-red/vacation-calculation/internal/infrastructure/exporter/exel"
)

type Container struct {
	// Infrastructure
	DB           *sqlx.DB
	Logger       ports.Logger
	Hasher       ports.PasswordHasher
	LoginUseCase *login.Handler

	// Repositories
	UserRepo       ports.UserRepository
	DistrictRepo   ports.DistrictRepository
	DepartmentRepo ports.DepartmentRepository
	// VacationRepo ports.VacationRepository

	// Use Cases - Commands (User)
	RegisterUserUseCase    *register_user.Handler
	BootstrapAdminUseCase  *bootstrap_admin.Handler
	DeactivateUserUseCase  *deactivate_user.Handler
	ActivateUserUseCase    *activate_user.Handler
	GetRegisterFormUseCase *register_form.Handler
	// Use Cases - Commands (Vacation)
	ApproveVacationUseCase      *approve_vacation.Handler
	ChangeVacationStatusUseCase *change_status.Handler

	CreateVacationUseCase   *create_vacations.Handler
	TransferVacationUseCase *transfer_vacation.Handler
	GetTransferFormUseCase  *get_transfer_form.Handler
	GetUserVacationsUseCase *get_user_vacations.Handler
	GetVacationFormUseCase  *get_vacation_form.Handler
	ExportReportUseCase     *export_report.Handler

	// Use Cases - Queries (User)
	GetUserUseCase            *get_user.Handler
	ListUsersUseCase          *list_users.Handler
	GetCalendarUseCase        *get_calendar.Handler
	UserCalendarUseCase       *get_user_calendar.Handler
	NewCalendarDayFormUseCase *new_calendar_day_form.Handler
	NewCalendarDayUseCase     *new_calendar_day.Handler

	// Уведомления о начале отпуска; nil — почта не настроена, рассылка выключена.
	SendRemindersUseCase *send_vacation_reminders.Handler

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
	var calendarDayRepo ports.CalendarDayRepository = sqlite.NewCalendarDayRepository(db)
	var transferRepo ports.TransferRepository = sqlite.NewTransferRepository(db)

	var exporter ports.ReportExporter = excel.NewReportExporter()

	registerUserUseCase := register_user.NewHandler(userRepo, districtRepo, hasher, log)
	deactivateUserUseCase := deactivate_user.NewHandler(userRepo, log)
	activateUserUseCase := activate_user.NewHandler(userRepo, log)
	getRegisterFormUseCase := register_form.NewHandler(districtRepo, departmentRepo)

	bootstrapAdminUseCase := bootstrap_admin.NewHandler(userRepo, registerUserUseCase)

	getUserUseCase := get_user.NewHandler(userRepo, vacationRepo, shiftRepo, calendarDayRepo, wc, log)
	getActiveUsersUseCase := list_users.NewHandler(userRepo, log)

	createVacationUseCase := create_vacations.NewHandler(vacationRepo, userRepo, shiftRepo, calendarDayRepo, wc)
	approveVacationUseCase := approve_vacation.NewHandler(vacationRepo, userRepo)
	changeVacationStatusUseCase := change_status.NewHandler(vacationRepo)
	transferVacationUseCase := transfer_vacation.NewHandler(vacationRepo, transferRepo, createVacationUseCase)
	getTransferFormUseCase := get_transfer_form.NewHandler(vacationRepo)
	getUserVacationsUseCase := get_user_vacations.NewHandler(vacationRepo, transferRepo)
	getVacationFormUseCase := get_vacation_form.NewHandler(userRepo, vacationRepo, shiftRepo, calendarDayRepo, wc)
	getCalendarUseCase := get_calendar.NewHandler(calendarDayRepo)
	userCalendarUseCase := get_user_calendar.NewHandler(calendarDayRepo, vacationRepo)
	newCalendarDayFormUseCase := new_calendar_day_form.NewHandler(calendarDayRepo)
	exportReportUseCase := export_report.NewHandler(userRepo, vacationRepo, departmentRepo, districtRepo, wc, exporter)

	newCalendarDayUseCase := new_calendar_day.NewHandler(calendarDayRepo)

	var sendRemindersUseCase *send_vacation_reminders.Handler
	if cfg.SMTPHost != "" && cfg.SMTPUsername != "" && cfg.SMTPPassword != "" {
		smtpMailer, err := mailer.NewSMTPMailer(mailer.Config{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUsername,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
			FromName: cfg.SMTPFromName,
			TLSMode:  cfg.SMTPTLS,
		})
		if err != nil {
			return nil, fmt.Errorf("mailer: %w", err)
		}
		// Отдел кадров: HR_EMAIL (через запятую) или ящик администратора.
		hrEmails := splitEmails(cfg.HREmail)
		if len(hrEmails) == 0 {
			hrEmails = []string{smtpMailer.From()}
		}
		for _, e := range hrEmails {
			if _, err := mail.ParseAddress(e); err != nil {
				return nil, fmt.Errorf("HR_EMAIL: invalid address %q: %w", e, err)
			}
		}
		reminderRepo := sqlite.NewReminderRepository(db)
		sendRemindersUseCase = send_vacation_reminders.NewHandler(vacationRepo, userRepo, reminderRepo, smtpMailer, hrEmails, log)
	}

	createShiftUseCase := create_shift.NewHandler(shiftRepo)
	shiftFormUseCase := shift_form.NewHandler()

	return &Container{
		// Infrastructure
		DB:           db,
		Logger:       log,
		Hasher:       hasher,
		LoginUseCase: login.NewHandler(userRepo, hasher),
		// Repositories
		UserRepo: userRepo,
		// Use Cases - Commands
		RegisterUserUseCase:    registerUserUseCase,
		BootstrapAdminUseCase:  bootstrapAdminUseCase,
		DeactivateUserUseCase:  deactivateUserUseCase,
		ActivateUserUseCase:    activateUserUseCase,
		GetRegisterFormUseCase: getRegisterFormUseCase,
		// Use Cases - Queries
		GetUserUseCase:              getUserUseCase,
		ListUsersUseCase:            getActiveUsersUseCase,
		DistrictRepo:                districtRepo,
		DepartmentRepo:              departmentRepo,
		CreateVacationUseCase:       createVacationUseCase,
		TransferVacationUseCase:     transferVacationUseCase,
		GetTransferFormUseCase:      getTransferFormUseCase,
		ApproveVacationUseCase:      approveVacationUseCase,
		ChangeVacationStatusUseCase: changeVacationStatusUseCase,
		GetUserVacationsUseCase:     getUserVacationsUseCase,
		GetVacationFormUseCase:      getVacationFormUseCase,
		GetCalendarUseCase:          getCalendarUseCase,
		UserCalendarUseCase:         userCalendarUseCase,
		NewCalendarDayFormUseCase:   newCalendarDayFormUseCase,
		NewCalendarDayUseCase:       newCalendarDayUseCase,
		CreateShiftUseCase:          createShiftUseCase,
		SendRemindersUseCase:        sendRemindersUseCase,
		ShiftFormUseCase:            shiftFormUseCase,
		ExportReportUseCase:         exportReportUseCase,
	}, nil
}

func splitEmails(s string) []string {
	out := make([]string, 0)
	for _, e := range strings.Split(s, ",") {
		if e = strings.TrimSpace(e); e != "" {
			out = append(out, e)
		}
	}
	return out
}

func (c *Container) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

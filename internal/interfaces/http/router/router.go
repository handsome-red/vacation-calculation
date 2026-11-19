package routes

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/di"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/handlers"
	authmiddleware "github.com/handsome-red/vacation-calculation/internal/interfaces/http/middleware"
	"github.com/handsome-red/vacation-calculation/web"
)

type Router struct {
	container    *di.Container
	logger       ports.Logger
	authMW       func(http.Handler) http.Handler
	sessions     *authmiddleware.AuthMiddleware
	loginHandler *handlers.LoginHandler

	templates           *handlers.Templates
	healthHandler       *handlers.HealthHandler
	homeHandler         *handlers.Handler
	registrationHandler *handlers.RegistrationHandler
	userHandler         *handlers.UserHandler
	vacationHandler     *handlers.VacationHandler
	shiftHandler        *handlers.ShiftHandler
	workYearHandler     *handlers.WorkYearHandler
	transferHandler     *handlers.TransferHandler
	scheduleHandler     *handlers.ScheduleHandler
	calendarHandler     *handlers.CalendarHandler
	auditHandler        *handlers.AuditHandler
	discrepancies       *handlers.DiscrepanciesHandler
	auditor             *authmiddleware.Auditor
}

func NewRouter(
	container *di.Container,
	logger ports.Logger,
	authMW func(http.Handler) http.Handler,
) (*Router, error) {
	templates, err := handlers.NewTemplates()
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	r := &Router{
		container: container,
		logger:    logger,
		authMW:    authMW,
		templates: templates,
	}

	r.healthHandler = handlers.NewHealthHandler()
	sessions := authmiddleware.NewAuthMiddleware(container.UserRepo, container.SessionRepo)
	r.authMW = sessions.Require
	r.sessions = sessions
	r.loginHandler = handlers.NewLoginHandler(container.LoginUseCase, sessions, authmiddleware.NewLoginLimiter(), templates)
	r.homeHandler = handlers.NewHomeHandler(templates)
	r.registrationHandler = handlers.NewRegistrationHandler(
		container.RegisterUserUseCase,
		container.GetRegisterFormUseCase,
		templates,
	)
	r.userHandler = handlers.NewUserHandler(
		container.GetUserUseCase,
		container.ListUsersUseCase,
		container.DeactivateUserUseCase,
		container.ActivateUserUseCase,
		container.UpdateUserUseCase,
		container.UserEditFormUseCase,
		templates,
		logger,
	)
	r.vacationHandler = handlers.NewVacationHandler(
		container.CreateVacationUseCase,
		container.ApproveVacationUseCase,
		container.ChangeVacationStatusUseCase,
		container.DeleteVacationUseCase,
		container.GetUserVacationsUseCase,
		container.GetVacationFormUseCase,
		container.ExportReportUseCase,
		templates,
		logger,
	)
	r.scheduleHandler = handlers.NewScheduleHandler(
		container.GetScheduleUseCase,
		container.ApproveScheduleUseCase,
		container.ScheduleRenderer,
		templates,
		logger,
	)
	r.transferHandler = handlers.NewTransferHandler(
		container.TransferVacationUseCase,
		container.GetTransferFormUseCase,
		templates,
		logger,
	)
	r.shiftHandler = handlers.NewShiftHandler(
		container.CreateShiftUseCase,
		container.ShiftFormUseCase,
		templates,
	)
	r.workYearHandler = handlers.NewWorkYearHandler(
		container.WorkYearsUseCase,
		container.SetWorkYearUseCase,
		templates,
		logger,
	)
	r.calendarHandler = handlers.NewCalendarHandler(
		container.GetCalendarUseCase,
		container.UserCalendarUseCase,
		container.NewCalendarDayUseCase,
		container.NewCalendarDayFormUseCase,
		templates,
		logger,
	)
	r.auditHandler = handlers.NewAuditHandler(container.ListAuditUseCase, templates, logger)
	r.discrepancies = handlers.NewDiscrepanciesHandler(container.ListDiscrepanciesUseCase, templates, logger)
	r.auditor = authmiddleware.NewAuditor(container.UserRepo, container.AuditRepo, logger, auditActions(container.VacationRepo))

	return r, nil
}

func (r *Router) Build() (http.Handler, error) {
	// Глобальный middleware для всего приложения
	root := chi.NewRouter()
	root.Use(middleware.RequestID)
	root.Use(middleware.RealIP)
	root.Use(middleware.Logger)
	root.Use(middleware.Recoverer)
	root.Use(authmiddleware.SameOrigin)

	// 1. Публичные маршруты (без аутентификации)
	r.registerPublic(root)

	// 2. API v1 (с аутентификацией)
	root.Route("/api/v1", func(api chi.Router) {
		if r.authMW != nil { // Middleware применяется ко всей группе /api/v1
			api.Use(r.authMW)
		}
		api.Use(r.auditor.Middleware)
		r.registerAPI(api)
	})

	// 3. Админка (аутентификация + проверка роли: администратор или руководитель)
	root.Route("/admin", func(admin chi.Router) {
		if r.authMW != nil {
			admin.Use(r.authMW)
		}
		admin.Use(r.auditor.Middleware)
		admin.Use(authmiddleware.RequireStaff)
		r.registerAdmin(admin)
	})

	return root, nil
}

func (r *Router) registerPublic(mux chi.Router) {
	// Static files
	staticFS, err := fs.Sub(web.FS, "static")
	if err != nil {
		panic("router: static fs: " + err.Error())
	}
	mux.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Public handlers (pre-initialized in NewRouter)
	mux.Get("/health", r.healthHandler.Health)
	mux.Get("/login", r.loginHandler.Form)
	mux.Post("/login", r.loginHandler.Login)
	mux.Post("/logout", r.loginHandler.Logout)

	mux.Group(func(private chi.Router) {
		private.Use(r.authMW)
		private.Use(r.auditor.Middleware)
		managerOf := r.sessions.RequireManagerOf("id")
		private.Get("/", r.homeHandler.Home)
		private.With(authmiddleware.RequireStaff).Get("/user/register", r.registrationHandler.Form)
		private.With(authmiddleware.RequireStaff).Post("/user/register", r.registrationHandler.Register)
		private.With(authmiddleware.RequireStaff).Get("/users", r.userHandler.ListUsersPage)
		private.With(r.sessions.RequireSelfOrManager("id")).Get("/users/{id}", r.userHandler.GetUserPage)
		private.With(managerOf).Get("/users/{id}/edit", r.userHandler.EditForm)
		private.With(managerOf).Post("/users/{id}/edit", r.userHandler.UpdateUser)
		private.With(managerOf).Get("/users/{id}/work-years", r.workYearHandler.Page)
		private.With(managerOf).Post("/users/{id}/work-years/{number}", r.workYearHandler.Set)
		private.With(managerOf).Get("/users/{id}/history", r.auditHandler.UserPage)
		private.Get("/calendar", r.calendarHandler.GetCalendarPage)
	})
}

func (r *Router) registerAPI(mux chi.Router) {
	// Здесь все пути БЕЗ префикса /api/v1, потому что мы уже внутри Route("/api/v1")

	// Reuse pre-initialized handlers from the router
	// Сотрудник — сам себе; администратор и руководитель его теротдела — как администратор.
	selfOrManager := r.sessions.RequireSelfOrManager("userId")
	managerOf := r.sessions.RequireManagerOf("userId")

	if r.userHandler != nil {
		mux.With(r.sessions.RequireSelfOrManager("id")).Get("/users/{id}", r.userHandler.GetUser)
		mux.With(r.sessions.RequireManagerOf("id")).Delete("/users/{id}", r.userHandler.DeactivateUser)
		mux.With(authmiddleware.RequireStaff).Get("/users", r.userHandler.ListUsers)
	}

	if r.vacationHandler != nil {
		mux.With(selfOrManager).Get("/users/{userId}/vacations", r.vacationHandler.GetVacations)
		mux.With(selfOrManager).Get("/users/{userId}/vacations/new", r.vacationHandler.GetVacationForm)
		mux.With(selfOrManager).Post("/users/{userId}/vacations", r.vacationHandler.CreateVacation)
		mux.With(selfOrManager).Post("/users/{userId}/vacations/preview", r.vacationHandler.PreviewVacation)
		mux.With(selfOrManager).Post("/users/{userId}/vacations/limit", r.vacationHandler.VacationLimit)
		mux.With(managerOf).Post("/users/{userId}/vacations/{id}/approve", r.vacationHandler.ApproveVacation)
		mux.With(managerOf).Post("/users/{userId}/vacations/{id}/status", r.vacationHandler.ChangeVacationStatus)
		mux.With(managerOf).Post("/users/{userId}/vacations/{id}/delete", r.vacationHandler.DeleteVacation)
		mux.With(managerOf).Get("/users/{userId}/vacations/{id}/transfer", r.transferHandler.Form)
		mux.With(managerOf).Post("/users/{userId}/vacations/{id}/transfer", r.transferHandler.Transfer)
	}

	if r.calendarHandler != nil {
		mux.With(selfOrManager).Get("/users/{userId}/calendar", r.calendarHandler.GetUserCalendar)
	}

	if r.calendarHandler != nil {
		mux.Get("/calendar", r.calendarHandler.GetCalendar)
	}

	if r.shiftHandler != nil {
		mux.With(managerOf).Get("/users/{userId}/shift/new", r.shiftHandler.NewShiftForm)
		mux.With(managerOf).Post("/users/{userId}/shift", r.shiftHandler.CreateShift)
	}
}

func (r *Router) registerAdmin(mux chi.Router) {
	// Производственный календарь общий для всех теротделов — только администратор.
	mux.With(authmiddleware.RequireAdmin).Get("/holidays", r.calendarHandler.GetCalendarDayForm)
	mux.With(authmiddleware.RequireAdmin).Post("/holidays", r.calendarHandler.NewCalendarDay)
	mux.With(authmiddleware.RequireAdmin).Post("/holidays/delete", r.calendarHandler.DeleteCalendarDay)
	mux.With(authmiddleware.RequireAdmin).Post("/holidays/fill", r.calendarHandler.FillHolidays)

	// График — в пределах теротделов руководителя (проверяет обработчик).
	mux.Get("/schedule", r.scheduleHandler.Page)
	mux.Post("/schedule/approve", r.scheduleHandler.Approve)

	// Журнал изменений — в пределах теротделов руководителя.
	mux.Get("/audit", r.auditHandler.Page)
	mux.Get("/discrepancies", r.discrepancies.Page)
	mux.Get("/reports/schedule", r.scheduleHandler.Report)
}

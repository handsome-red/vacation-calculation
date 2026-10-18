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
	loginHandler *handlers.LoginHandler

	templates           *handlers.Templates
	healthHandler       *handlers.HealthHandler
	homeHandler         *handlers.Handler
	registrationHandler *handlers.RegistrationHandler
	userHandler         *handlers.UserHandler
	vacationHandler     *handlers.VacationHandler
	shiftHandler        *handlers.ShiftHandler
	calendarHandler     *handlers.CalendarHandler
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
	sessions := authmiddleware.NewAuthMiddleware(container.UserRepo)
	r.authMW = sessions.Require
	r.loginHandler = handlers.NewLoginHandler(container.LoginUseCase, sessions, templates)
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
		templates,
		logger,
	)
	r.vacationHandler = handlers.NewVacationHandler(
		container.CreateVacationUseCase,
		container.ApproveVacationUseCase,
		container.ChangeVacationStatusUseCase,
		container.GetUserVacationsUseCase,
		container.GetVacationFormUseCase,
		container.ExportReportUseCase,
		templates,
		logger,
	)
	r.shiftHandler = handlers.NewShiftHandler(
		container.CreateShiftUseCase,
		container.ShiftFormUseCase,
		templates,
	)
	r.calendarHandler = handlers.NewCalendarHandler(
		container.GetCalendarUseCase,
		container.NewCalendarDayUseCase,
		container.NewCalendarDayFormUseCase,
		templates,
		logger,
	)

	return r, nil
}

func (r *Router) Build() (http.Handler, error) {
	// Глобальный middleware для всего приложения
	root := chi.NewRouter()
	root.Use(middleware.RequestID)
	root.Use(middleware.RealIP)
	root.Use(middleware.Logger)
	root.Use(middleware.Recoverer)

	// 1. Публичные маршруты (без аутентификации)
	r.registerPublic(root)

	// 2. API v1 (с аутентификацией)
	root.Route("/api/v1", func(api chi.Router) {
		// if r.authMW != nil { // Middleware применяется ко всей группе /api/v1
		// 	api.Use(r.authMW)
		// }
		r.registerAPI(api)
	})

	// 3. Админка (аутентификация + проверка роли)
	root.Route("/admin", func(admin chi.Router) {
		// if r.authMW != nil {
		// 	admin.Use(r.authMW)
		// }
		// admin.Use(adminRoleMW) // Если нужна проверка роли
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
		private.Get("/", r.homeHandler.Home)
		private.With(authmiddleware.RequireAdmin).Get("/user/register", r.registrationHandler.Form)
		private.With(authmiddleware.RequireAdmin).Post("/user/register", r.registrationHandler.Register)
		private.With(authmiddleware.RequireAdmin).Get("/users", r.userHandler.ListUsersPage)
		private.Get("/users/{id}", r.userHandler.GetUserPage)
		private.Get("/calendar", r.calendarHandler.GetCalendarPage)
	})
}

func (r *Router) registerAPI(mux chi.Router) {
	// Здесь все пути БЕЗ префикса /api/v1, потому что мы уже внутри Route("/api/v1")

	// Reuse pre-initialized handlers from the router
	if r.userHandler != nil {
		mux.Get("/users/{id}", r.userHandler.GetUser)
		mux.With(authmiddleware.RequireAdmin).Delete("/users/{id}", r.userHandler.DeactivateUser)
		mux.With(authmiddleware.RequireAdmin).Get("/users", r.userHandler.ListUsers)
	}

	if r.vacationHandler != nil {
		mux.Get("/users/{userId}/vacations", r.vacationHandler.GetVacations)
		mux.Get("/users/{userId}/vacations/new", r.vacationHandler.GetVacationForm)
		mux.Post("/users/{userId}/vacations", r.vacationHandler.CreateVacation)
		mux.With(authmiddleware.RequireAdmin).Post("/users/{userId}/vacations/{id}/approve", r.vacationHandler.ApproveVacation)
		mux.With(authmiddleware.RequireAdmin).Post("/users/{userId}/vacations/{id}/status", r.vacationHandler.ChangeVacationStatus)
	}

	if r.calendarHandler != nil {
		mux.Get("/users/{userId}/calendar", r.calendarHandler.GetUserCalendar)
	}

	if r.calendarHandler != nil {
		mux.Get("/calendar", r.calendarHandler.GetCalendar)
	}

	if r.shiftHandler != nil {
		mux.With(authmiddleware.RequireAdmin).Get("/users/{userId}/shift/new", r.shiftHandler.NewShiftForm)
		mux.With(authmiddleware.RequireAdmin).Post("/users/{userId}/shift", r.shiftHandler.CreateShift)
	}
}

func (r *Router) registerAdmin(mux chi.Router) {
	mux.Get("/holidays", r.calendarHandler.NewCalendarDay)
	mux.Post("/holidays", r.calendarHandler.NewCalendarDay)
}

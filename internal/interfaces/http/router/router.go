package routes

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/di"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/handlers"
	"github.com/handsome-red/vacation-calculation/web"
)

type Router struct {
	container *di.Container
	logger    ports.Logger
	authMW    func(http.Handler) http.Handler

	templates           *handlers.Templates
	healthHandler       *handlers.HealthHandler
	homeHandler         *handlers.HomeHandler
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
	r.homeHandler = handlers.NewHomeHandler(templates)
	r.registrationHandler = handlers.NewRegistrationHandler(
		container.RegisterUserUseCase,
		container.GetRegisterFormUseCase,
		templates,
	)
	r.userHandler = handlers.NewUserHandler(
		container.GetUserUseCase,
		container.DeactivateUserUseCase,
		container.ActivateUserUseCase,
		container.ListUsersUseCase,
		templates,
	)
	r.vacationHandler = handlers.NewVacationHandler(
		container.CreateVacationUseCase,
		container.GetUserVacationsUseCase,
		container.GetVacationFormUseCase,
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
		container.NewHolidayUseCase,
		container.NewHolidayFormUseCase,
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
		api.Use(r.authMW) // Middleware применяется ко всей группе /api/v1
		r.registerAPI(api)
	})

	// 3. Админка (аутентификация + проверка роли)
	root.Route("/admin", func(admin chi.Router) {
		admin.Use(r.authMW)
		// admin.Use(adminRoleMW) // Если нужна проверка роли
		r.registerAdmin(admin)
	})

	return root, nil
}

func (r *Router) registerPublic(mux chi.Router) {
	staticFS, err := fs.Sub(web.FS, "static")
	if err != nil {
		panic("router: static fs: " + err.Error())
	}
	mux.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	templates, err := handlers.NewTemplates()
	if err != nil {
		panic("router: templates: " + err.Error())
	}

	healthHandler := handlers.NewHealthHandler()
	homeHandler := handlers.NewHomeHandler(templates)

	mux.Get("/", homeHandler.Home)
	mux.Get("/health", healthHandler.Health)

	// Публичный регистрация
	userHandler := handlers.NewUserHandler(
		r.container.RegisterUserUseCase,
		r.container.GetUserUseCase,
		r.container.DeactivateUserUseCase,
		r.container.ActivateUserUseCase,
		r.container.ListUsersUseCase,
		r.container.GetRegisterFormUseCase,
		r.container.CreateShiftUseCase,
		r.container.ShiftFormUseCase,
		r.teplates,
	)
	mux.Get("/user/register", userHandler.RegisterUserForm)
	mux.Post("/user/register", userHandler.RegisterUser)
}

func (r *Router) registerAPI(mux chi.Router) {
	// Здесь все пути БЕЗ префикса /api/v1, потому что мы уже внутри Route("/api/v1")
	userHandler := handlers.NewUserHandler( /* ... */ )
	vacationHandler := handlers.NewVacationHandler( /* ... */ )
	calendarHandler := handlers.NewCalendarHandler( /* ... */ )
	shiftHandler := handlers.NewShiftHandler( /* ... */ )

	mux.Get("/users/{id}", userHandler.GetUser)
	mux.Delete("/users/{id}", userHandler.DeactivateUser)
	mux.Get("/users", userHandler.ListUsers)

	mux.Get("/users/{userId}/vacations", vacationHandler.GetVacations)
	mux.Get("/users/{userId}/vacations/new", vacationHandler.GetVacationForm)
	mux.Post("/users/{userId}/vacations", vacationHandler.CreateVacation)

	mux.Get("/users/{userId}/calendar", calendarHandler.GetUserCalendar)

	mux.Get("/users/{userId}/shift/new", shiftHandler.NewShiftForm)
	mux.Post("/users/{userId}/shift", shiftHandler.CreateShift)
}

func (r *Router) registerAdmin(mux chi.Router) {
	calendarHandler := handlers.NewCalendarHandler( /* ... */ )
	mux.Get("/holidays", calendarHandler.NewHoliday)
	mux.Post("/holidays", calendarHandler.NewHoliday)
}

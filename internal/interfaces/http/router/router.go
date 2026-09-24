package routes

import (
	"io/fs"
	"net/http"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/di"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/handlers"
	"github.com/handsome-red/vacation-calculation/web"
)

type Router struct {
	mux         *http.ServeMux
	container   *di.Container
	logger      ports.Logger
	middlewares []func(http.Handler) http.Handler
	handler     http.Handler
}

func NewRouter(container *di.Container, logger ports.Logger) *Router {
	return &Router{
		mux:         http.NewServeMux(),
		container:   container,
		logger:      logger,
		middlewares: []func(http.Handler) http.Handler{},
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r.handler == nil {
		panic("routes.Router: Setup() must be called before ServeHTTP")
	}
	r.handler.ServeHTTP(w, req)
}

func (r *Router) WithMiddleware(mw ...func(http.Handler) http.Handler) *Router {
	r.middlewares = append(r.middlewares, mw...)
	return r
}

func (r *Router) Build() http.Handler {
	r.registerRoutes()

	handler := http.Handler(r.mux)

	for _, mw := range r.middlewares {
		handler = mw(handler)
	}

	return handler
}

func (r *Router) registerRoutes() {

	staticFS, err := fs.Sub(web.FS, "static")
	if err != nil {
		panic("router: static fs: " + err.Error())
	}
	r.mux.Handle(
		"GET /static/",
		http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))),
	)
	templates, err := handlers.NewTemplates()
	if err != nil {
		panic("router: templates: " + err.Error())
	}

	healthHandler := handlers.NewHealthHandler()

	userHandler := handlers.NewUserHandler(
		r.container.RegisterUserUseCase,
		r.container.GetUserUseCase,
		r.container.DeactivateUserUseCase,
		r.container.ActivateUserUseCase,
		r.container.GetActiveUsersUseCase,
		r.container.GetRegisterFormUseCase,
		// r.logger,
		templates,
	)

	vacationHandler := handlers.NewVacationHandler(
		r.container.CreateVacationUseCase,
		r.container.GetUserVacationsUseCase,
		r.container.GetVacationFormUseCase,
		r.logger,
	)

	// calendarHandler := handlers.NewCalendarHandlers(
	// 	r.container.
	// )

	r.mux.HandleFunc("GET /{$}", homeHandler)

	// ============================================================
	// Public Routes (без аутентификации)
	// ============================================================
	r.mux.HandleFunc("GET /health", healthHandler.Health)
	// r.mux.HandleFunc("GET /ready", healthHandler.Ready)

	// ============================================================
	// API v1 Routes
	// ============================================================
	// User routes
	// TODO подумать над маршрутом для html
	r.mux.HandleFunc("GET /api/v1/user/register", userHandler.RegisterUserForm)
	r.mux.HandleFunc("POST /api/v1/user/register", userHandler.RegisterUser)

	r.mux.HandleFunc("GET /api/v1/users/{id}", userHandler.GetUser)
	// r.mux.HandleFunc("PUT /api/v1/users/{id}/email", userHandler.ChangeEmail)
	r.mux.HandleFunc("DELETE /api/v1/users/{id}", userHandler.DeactivateUser)
	// r.mux.HandleFunc("POST /api/v1/users/{id}/activate", userHandler.ActivateUser)
	r.mux.HandleFunc("GET /api/v1/users", userHandler.GetActiveUsers)

	// Calendar
	// r.mux.HandleFunc("GET /api/v1/calendar", calendarHandler.GetCalendar)

	// Vacation routes
	// r.mux.HandleFunc("POST /api/v1/vacations", vacationHandler.CreateVacation)
	r.mux.HandleFunc("GET /api/v1/users/{userId}/vacations", vacationHandler.GetVacation)
	r.mux.HandleFunc("POST /api/v1/users/{userId}/vacations", vacationHandler.SetVacation)
	// r.mux.HandleFunc("GET /api/v1/users/{userId}/vacations", vacationHandler.GetUserVacations)
	// r.mux.HandleFunc("POST /api/v1/vacations/{id}/approve", vacationHandler.ApproveVacation)
	// r.mux.HandleFunc("POST /api/v1/vacations/{id}/reject", vacationHandler.RejectVacation)
	// r.mux.HandleFunc("DELETE /api/v1/vacations/{id}", vacationHandler.CancelVacation)
}

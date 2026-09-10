package app

import (
	"net/http"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/di"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/handlers"
)

type Router struct {
	mux         *http.ServeMux
	container   *di.Container
	logger      ports.Logger
	middlewares []func(http.Handler) http.Handler
}

func NewRouter(container *di.Container, logger ports.Logger) *Router {
	return &Router{
		mux:         http.NewServeMux(),
		container:   container,
		logger:      logger,
		middlewares: []func(http.Handler) http.Handler{},
	}
}

func (r *Router) WithMiddleware(mw ...func(http.Handler) http.Handler) *Router {
	r.middlewares = append(r.middlewares, mw...)
	return r
}

func (r *Router) Build() http.Handler {
	r.registerRoutes()

	handler := http.Handler(r.mux)

	for i := len(r.middlewares); i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}

	return handler
}

func (r *Router) registerRoutes() {
	healthHandler := handlers.NewHealthHandler(r.logger)
	userHandler := handlers.NewUserHandler(
		r.container.RegisterUserUseCase,
		r.container.GetUserUseCase,
		r.container.DeactivateUserUseCase,
		r.container.ActivateUserUseCase,
		r.container.GetActiveUsersUseCase,
		r.logger,
	)
	vacationHandler := handlers.NewVacationHandler(
		r.container.CreateVacationUseCase,
		r.container.ApproveVacationUseCase,
		r.container.GetUserVacationsUseCase,
		r.logger,
	)
	// ============================================================
	// Public Routes (без аутентификации)
	// ============================================================
	r.mux.HandleFunc("GET /health", healthHandler.Health)
	r.mux.HandleFunc("GET /ready", healthHandler.Ready)

	// ============================================================
	// API v1 Routes
	// ============================================================
	// User routes
	r.mux.HandleFunc("POST /api/v1/users", userHandler.RegisterUser)
	r.mux.HandleFunc("GET /api/v1/users/{id}", userHandler.GetUser)
	r.mux.HandleFunc("PUT /api/v1/users/{id}/email", userHandler.ChangeEmail)
	r.mux.HandleFunc("DELETE /api/v1/users/{id}", userHandler.DeactivateUser)
	r.mux.HandleFunc("POST /api/v1/users/{id}/activate", userHandler.ActivateUser)
	r.mux.HandleFunc("GET /api/v1/users", userHandler.GetActiveUsers)

	// Vacation routes
	r.mux.HandleFunc("POST /api/v1/vacations", vacationHandler.CreateVacation)
	r.mux.HandleFunc("GET /api/v1/vacations/{id}", vacationHandler.GetVacation)
	r.mux.HandleFunc("GET /api/v1/users/{userId}/vacations", vacationHandler.GetUserVacations)
	r.mux.HandleFunc("POST /api/v1/vacations/{id}/approve", vacationHandler.ApproveVacation)
	r.mux.HandleFunc("POST /api/v1/vacations/{id}/reject", vacationHandler.RejectVacation)
	r.mux.HandleFunc("DELETE /api/v1/vacations/{id}", vacationHandler.CancelVacation)
}

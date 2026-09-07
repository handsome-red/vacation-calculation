package app

import (
	"log/slog"
	"net/http"

	"github.com/handsome-red/vacation-calculation/internal/handler"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/api/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(hand *handler.Handler, log *slog.Logger) *chi.Mux {
	mux := chi.NewRouter()

	mux.Use(httplog.Logger(log))
	mux.Use(middleware.Recoverer)

	mux.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// HTML страницы
	mux.Get("/health", hand.HealthHandler) // Здоровье
	mux.Get("/ready", hand.ReadyHandler)   // Готовность

	mux.Get("/", hand.IndexHandler)             // Главная
	mux.Get("/vacations", hand.ListHandler)     // Список
	mux.Get("/add", hand.AddFormHandler)        // Форма добавления
	mux.Post("/add", hand.CreateHandler)        // Добавление
	mux.Get("/edit/{id}", hand.EditFormHandler) // Форма редактирования
	mux.Post("/edit/{id}", hand.UpdateHandler)  // Обновление
	mux.Get("/delete/{id}", hand.DeleteHandler) // Удаление
	mux.Get("/view/{id}", hand.ViewHandler)     // Просмотр

	// API
	mux.Post("/api/calculate", hand.CalculateHandler)

	return mux
}

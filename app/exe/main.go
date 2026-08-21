package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"vacation-calculation/internal/handler"
	"vacation-calculation/internal/repository"
	"vacation-calculation/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	templates, err := loadTemplates()
	if err != nil {
		logger.Error("Failed to load templates", "error", err)
		os.Exit(1)
	}

	repo, err := repository.NewSQLiteRepository("./data/vacations.db")
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return
	}

	serv := service.NewVacationService(repo)
	hand := handler.NewHandler(serv, templates)

	mux := chi.NewRouter()

	mux.Use(middleware.Logger)
	mux.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// HTML страницы
	mux.Get("/", hand.IndexHandler)              // Главная
	mux.Get("/vacations", hand.ListHandler)      // Список
	mux.Get("/add", hand.AddFormHandler)         // Форма добавления
	mux.Post("/add", hand.CreateHandler)         // Добавление
	mux.Get("/edit/{id}", hand.EditFormHandler)  // Форма редактирования
	mux.Post("/edit/{id}", hand.UpdateHandler)   // Обновление
	mux.Get("/delete/{id}", hand.DeleteHandler)  // Удаление 
	mux.Get("/view/{id}", hand.ViewHandler)      // Просмотр

	// API
	mux.Post("/api/calculate", hand.CalculateHandler)


	addr := ":8080"

	logger.Info("Server is running", "addr", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

// app/exe/main.go
func loadTemplates() (*template.Template, error) {
    // Просто загружаем все файлы из одной папки
    return template.ParseGlob("web/templates/*.html")
}
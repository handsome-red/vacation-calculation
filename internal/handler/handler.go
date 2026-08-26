package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"
	"vacation-calculation/internal/service"
)

type Handler struct {
	service   service.VacationService
	templates *template.Template
}

func NewHandler(service service.VacationService, templates *template.Template) *Handler {
	return &Handler{
		service:   service,
		templates: templates,
	}
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadyHandler проверяет готовность приложения
func (h *Handler) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем БД
	if err := h.service.Ping(); err != nil {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {

	vacations, err := h.service.GetAllVacations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":     "Главная страница",
		"Message":   "Добро пожаловать",
		"Vacations": vacations,
	}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	vacations, err := h.service.GetAllVacations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":     "Главная страница",
		"Message":   "Добро пожаловать",
		"Vacations": vacations,
	}

	if err := h.templates.ExecuteTemplate(w, "vacations.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

func (h *Handler) AddFormHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.templates.ExecuteTemplate(w, "vacation_form.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) EditFormHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

func (h *Handler) ViewHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

func (h *Handler) CalculateHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

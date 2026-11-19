package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/set_work_year"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/work_years"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	authmiddleware "github.com/handsome-red/vacation-calculation/internal/interfaces/http/middleware"
)

type WorkYearHandler struct {
	view      *work_years.Handler
	set       *set_work_year.Handler
	templates *Templates
	logger    ports.Logger
}

func NewWorkYearHandler(view *work_years.Handler, set *set_work_year.Handler, templates *Templates, logger ports.Logger) *WorkYearHandler {
	return &WorkYearHandler{view: view, set: set, templates: templates, logger: logger}
}

// Page — границы рабочих лет сотрудника (только администратор).
func (h *WorkYearHandler) Page(w http.ResponseWriter, r *http.Request) {
	view, err := h.view.Handle(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.renderError(w, r, err)
		return
	}
	if err := h.templates.Render(w, "work_years.html", view); err != nil {
		h.logger.Error(r.Context(), "render work years page failed", "error", err)
	}
}

// Set сохраняет границы года; при ошибке ввода показывает страницу с текстом ошибки.
func (h *WorkYearHandler) Set(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	id := chi.URLParam(r, "id")
	number, err := strconv.Atoi(chi.URLParam(r, "number"))
	if err != nil {
		http.Error(w, "invalid work year number", http.StatusBadRequest)
		return
	}

	err = h.set.Handle(r.Context(), set_work_year.Command{
		UserID: id,
		Number: number,
		Start:  r.FormValue("start"),
		End:    r.FormValue("end"),
		Reset:  r.FormValue("action") == "reset",
	})
	if err == nil {
		admin, _ := authmiddleware.UserFromContext(r.Context())
		h.logger.Info(r.Context(), "work year changed", "user_id", id, "number", number, "admin_id", admin.ID().String())
		http.Redirect(w, r, "/vacations/users/"+id+"/work-years", http.StatusSeeOther)
		return
	}
	if !errors.Is(err, set_work_year.ErrInvalidInput) {
		h.renderError(w, r, err)
		return
	}

	view, verr := h.view.Handle(r.Context(), id)
	if verr != nil {
		h.renderError(w, r, verr)
		return
	}
	view.Error = err.Error()
	w.WriteHeader(http.StatusBadRequest)
	if err := h.templates.Render(w, "work_years.html", view); err != nil {
		h.logger.Error(r.Context(), "render work years page failed", "error", err)
	}
}

func (h *WorkYearHandler) renderError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, user.ErrUserNotFound) {
		http.Error(w, "Сотрудник не найден", http.StatusNotFound)
		return
	}
	h.logger.Error(r.Context(), "work years failed", "error", err)
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

package handlers

import (
	"net/http"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	authmiddleware "github.com/handsome-red/vacation-calculation/internal/interfaces/http/middleware"
)

type Handler struct {
	templates *Templates
}

func NewHomeHandler(templates *Templates) *Handler {
	return &Handler{
		templates: templates,
	}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{"IsAdmin": false, "UserID": ""}
	if u, ok := authmiddleware.UserFromContext(r.Context()); ok && u != nil {
		data["IsAdmin"] = u.IsAdmin()
		data["UserID"] = u.ID().String()
	}

	// С октября и до крайнего срока утверждения — предложить сформировать график на следующий год.
	now := time.Now()
	if year := ScheduleYear(now); year > now.Year() {
		deadline := vacation.ApprovalDeadline(year)
		if !vacation.DateOf(now).After(deadline) {
			data["ScheduleYear"] = year
			data["ScheduleDeadline"] = deadline.Time().Format("02.01.2006")
		}
	}

	if err := h.templates.Render(w, "home.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

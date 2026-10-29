package handlers

import (
	"net/http"

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

	if err := h.templates.Render(w, "home.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

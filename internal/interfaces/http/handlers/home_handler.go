package handlers

import "net/http"

type Handler struct {
	templates *Templates
}

func NewHomeHandler(templates *Templates) *Handler {
	return &Handler{
		templates: templates,
	}
}

func(h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if err := h.templates.Render(w, "home.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
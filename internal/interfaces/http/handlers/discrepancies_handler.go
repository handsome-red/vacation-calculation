package handlers

import (
	"net/http"

	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/list_discrepancies"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/middleware"
)

type DiscrepanciesHandler struct {
	list      *list_discrepancies.Handler
	templates *Templates
	logger    ports.Logger
}

func NewDiscrepanciesHandler(list *list_discrepancies.Handler, templates *Templates, logger ports.Logger) *DiscrepanciesHandler {
	return &DiscrepanciesHandler{list: list, templates: templates, logger: logger}
}

// Page — сотрудники с расхождениями в учёте отпусков (в пределах теротделов руководителя).
func (h *DiscrepanciesHandler) Page(w http.ResponseWriter, r *http.Request) {
	caller, ok := middleware.CallerFromContext(r.Context())
	if !ok {
		http.Error(w, "Необходим вход", http.StatusUnauthorized)
		return
	}
	result, err := h.list.Handle(r.Context(), list_discrepancies.Query{Caller: caller})
	if err != nil {
		h.logger.Error(r.Context(), "list discrepancies failed", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if err := h.templates.Render(w, "discrepancies.html", result); err != nil {
		h.logger.Error(r.Context(), "render discrepancies page failed", "error", err)
	}
}

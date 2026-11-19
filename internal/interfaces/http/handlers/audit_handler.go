package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/handsome-red/vacation-calculation/internal/application/queries/audit/list_entries"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/middleware"
)

type AuditHandler struct {
	list      *list_entries.Handler
	templates *Templates
	logger    ports.Logger
}

func NewAuditHandler(list *list_entries.Handler, templates *Templates, logger ports.Logger) *AuditHandler {
	return &AuditHandler{list: list, templates: templates, logger: logger}
}

// Page — общий журнал изменений (администратор и руководители, в пределах своих теротделов).
func (h *AuditHandler) Page(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "")
}

// UserPage — журнал изменений по одному сотруднику.
func (h *AuditHandler) UserPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, chi.URLParam(r, "id"))
}

func (h *AuditHandler) render(w http.ResponseWriter, r *http.Request, targetID string) {
	caller, ok := middleware.CallerFromContext(r.Context())
	if !ok {
		http.Error(w, "Необходим вход", http.StatusUnauthorized)
		return
	}
	result, err := h.list.Handle(r.Context(), list_entries.Query{Caller: caller, TargetID: targetID})
	if err != nil {
		h.logger.Error(r.Context(), "list audit entries failed", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	data := struct {
		*list_entries.Result
		UserID string
	}{result, targetID}
	if err := h.templates.Render(w, "audit.html", data); err != nil {
		h.logger.Error(r.Context(), "render audit page failed", "error", err)
	}
}

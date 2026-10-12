// internal/interfaces/http/handlers/user.go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/activate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/deactivate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/get_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/list_users"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

type UserHandler struct {
	getUser    *get_user.Handler
	listUsers  *list_users.Handler
	deactivate *deactivate_user.Handler
	activate   *activate_user.Handler
	templates  *Templates
	logger     ports.Logger
}

func NewUserHandler(
	getUser *get_user.Handler,
	listUsers *list_users.Handler,
	deactivate *deactivate_user.Handler,
	activate *activate_user.Handler,
	templates *Templates,
	logger ports.Logger,
) *UserHandler {
	return &UserHandler{
		getUser:    getUser,
		listUsers:  listUsers,
		deactivate: deactivate,
		activate:   activate,
		templates:  templates,
		logger:     logger,
	}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	view, err := h.getUser.Handle(r.Context(), get_user.Query{UserID: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	h.writeJSON(w, http.StatusOK, view)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	view, err := h.listUsers.Handle(r.Context(), list_users.Query{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, http.StatusOK, view)
}

func (h *UserHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.deactivate.Handle(r.Context(), deactivate_user.Command{UserID: id}); err != nil {
		h.logger.Error(context.Background(), "deactivate user failed", "user_id", id.String(), "error", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.activate.Handle(r.Context(), activate_user.Command{UserID: id}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if payload == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Error(context.Background(), "writeJSON: encode failed", "error", err)
	}
}

func (h *UserHandler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}
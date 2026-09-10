package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/activate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/deactivate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/get_user"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/dto"
)

type UserHandler struct {
	registerUseCase   *register_user.Handler
	getUserUseCase    *get_user.Handler
	deactivateUseCase *deactivate_user.Handler
	activateUseCase   *activate_user.Handler
}

func NewUserHandler(
	registerUseCase *register_user.Handler,
	getUserUseCase *get_user.Handler,
	deactivateUseCase *deactivate_user.Handler,
	activateUseCase *activate_user.Handler,
	getActiveUsersUseCase *get_active_user.GetActiveUsersUseCase,
) *UserHandler {
	return &UserHandler{
		registerUseCase:   registerUseCase,
		getUserUseCase:    getUserUseCase,
		deactivateUseCase: deactivateUseCase,
		activateUseCase:   activateUseCase,
	}
}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {

	var req dto.RegisterUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	cmd := register_user.Command{
		Email:      req.Email,
		Password:   req.Password,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		MiddleName: req.MiddleName,
		Department: req.Department,
	}

	result, err := h.registerUseCase.Handle(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	query := get_user.Query{UserID: userID}
	result, err := h.getUserUseCase.Handle(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *UserHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	cmd := deactivate_user.Command{UserID: userID}
	if err := h.deactivateUseCase.Handle(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

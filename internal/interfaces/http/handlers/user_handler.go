package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/activate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/deactivate_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/get_active_users"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/get_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/register_form"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/dto"
)

type UserHandler struct {
	registerUseCase         *register_user.Handler
	getUserUseCase          *get_user.Handler
	deactivateUseCase       *deactivate_user.Handler
	activateUseCase         *activate_user.Handler
	getActiveUsersUseCase   *get_active_users.Handler
	registerFormUserUseCase *register_form.Handler
	templates               *Templates
}

func NewUserHandler(
	registerUseCase *register_user.Handler,
	getUserUseCase *get_user.Handler,
	deactivateUseCase *deactivate_user.Handler,
	activateUseCase *activate_user.Handler,
	getActiveUsersUseCase *get_active_users.Handler,
	registerFormUserUseCase *register_form.Handler,
	templates *Templates,
) *UserHandler {
	return &UserHandler{
		registerUseCase:         registerUseCase,
		getUserUseCase:          getUserUseCase,
		deactivateUseCase:       deactivateUseCase,
		activateUseCase:         activateUseCase,
		getActiveUsersUseCase:   getActiveUsersUseCase,
		registerFormUserUseCase: registerFormUserUseCase,
		templates:               templates,
	}
}

func (h *UserHandler) RegisterUserForm(w http.ResponseWriter, r *http.Request) {

	result, err := h.registerFormUserUseCase.Handle(r.Context(), register_form.Query{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Form":        register_user.Command{},
		"Positions":   result.Positions,
		"Districts":   result.Districts,
		"Departments": result.Departments,
	}

	if err := h.templates.Render(w, "register_user.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	cmd := register_user.Command{
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		FirstName:       r.FormValue("first_name"),
		LastName:        r.FormValue("last_name"),
		MiddleName:      r.FormValue("middle_name"),
		Status:          r.FormValue("status"),
		BirthDate:       r.FormValue("birth_date"),
		Position:        r.FormValue("position"),
		HiredAt:         r.FormValue("hired_at"),
		WorkdayDuration: atoiSafe(r.FormValue("workday_duration")),
		IsInvalid:       r.FormValue("is_invalid") == "on",
		DistrictCode:    r.FormValue("district"),
		DepartmentCode:  r.FormValue("department"),
	}

	log.Printf("user registered successfully, %v: ", cmd.Email)

	if _, err := h.registerUseCase.Handle(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	query := get_user.Query{UserID: userID}
	result, err := h.getUserUseCase.Handle(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// vacation, err := h.countVacation.Hanlde(r.Context())

	data := map[string]any{
		"User": result,
		// "Vacation": vacation,
	}

	if err := h.templates.Render(w, "user.html", data); err != nil {
		// h.logger.Error(r.Context(), "render user", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

// func (h *UserHandler) ChangeEmail(w http.ResponseWriter, r *http.Request) {
// 	userIDStr := r.PathValue("id")
// 	userID, err := uuid.Parse(userIDStr)
// 	if err != nil {
// 		http.Error(w, "invalid user ID", http.StatusBadRequest)
// 		return
// 	}

// 	cmd := change_email
// }

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

func (h *UserHandler) GetActiveUsers(w http.ResponseWriter, r *http.Request) {
	query := get_active_users.Query{
		Page: 1,
		Size: 1,
	}

	result, err := h.getActiveUsersUseCase.Handle(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := make([]dto.UserResponse, 0, len(result.Users))
	for _, u := range result.Users {
		resp = append(resp, dto.UserResponse{
			ID:              u.ID,
			FirstName:       u.FirstName,
			LastName:        u.LastName,
			MiddleName:      u.MiddleName,
			BirthDate:       u.BirthDate,
			Position:        u.Position,
			HiredAt:         u.HiredAt,
			DepartmentCode:  u.Department,
			DistrictCode:    u.District,
			WorkdayDuration: u.WorkdayDuration,
			Email:           u.Email,
			IsInvalid:       u.IsInvalid,
		})
	}

	data := map[string]any{
		"Users": resp,
	}

	if err := h.templates.Render(w, "users.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func atoiSafe(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

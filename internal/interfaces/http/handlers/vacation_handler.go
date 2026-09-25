package handlers

import (
	"net/http"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/create_vacation"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/get_vacation_form"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_user_vacations"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type VacationHandler struct {
	createVacationUseCase *create_vacation.Handler
	getUserVacationsUseCase *get_user_vacations.Handler
	getVacationFormUseCase  *get_vacation_form.Handler
	templates *Templates
}

func NewVacationHandler(
	createVacationUseCase *create_vacation.Handler,
	getUserVacationsUseCase *get_user_vacations.Handler,
	getVacationFormUseCase  *get_vacation_form.Handler,
	templates               *Templates,
) *VacationHandler {
	return &VacationHandler{
		createVacationUseCase:   createVacationUseCase,
		getUserVacationsUseCase: getUserVacationsUseCase,
		getVacationFormUseCase:  getVacationFormUseCase,
		templates:               templates,
	}
}

func (h *VacationHandler) GetVacations(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("user_id")
	userID, err := user.ParseUserID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result, err := h.getUserVacationsUseCase.Handle(r.Context(), get_user_vacations.Query{
		UserID: userID.String(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"UserID":   userID.String(),
		"Vacations": result.Items,
	}

	if err := h.templates.Render(w, "vacations.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *VacationHandler) CreateVacation(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	cmd := create_vacation.Command{
		UserID:    r.FormValue("user_id"),
		StartDate: r.FormValue("start_date"),
		EndDate:   r.FormValue("end_date"),
	}

	_, err := h.createVacationUseCase.Handle(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/vacations", http.StatusSeeOther)
}

func (h *VacationHandler) GetVacationForm(w http.ResponseWriter, r *http.Request) {
	
	userID := r.PathValue("userId")
	if _, err := user.ParseUserID(userID); err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	_, err := h.getVacationFormUseCase.Handle(r.Context(), get_vacation_form.Query{
		UserID: userID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Form":      create_vacation.Command{UserID: userID},
		"UserID":    userID,
	}

	if err := h.templates.Render(w, "vacation_form.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

package handlers

import (
	"net/http"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/create_vacation"
)

type VacationHandler struct {
	createVacationUseCase *create_vacation.Handler
	// getUserVacationsUseCase *get_user_vacations.Hanlder
	getVacationFormUseCase  *get_vacation_form.Handler
}

func NewVacationHandler(
	createVacationUseCase *create_vacation.Handler,
	getVacationFormUseCase  *get_vacation_form.Handler,
) *VacationHandler {
	return &VacationHandler{
		createVacationUseCase:  createVacationUseCase,
		getVacationFormUseCase: getVacationFormUseCase,
	}
}

func (h *VacationHandler) CreateVacation(w http.ResponseWriter, r *http.Request) {
	if err := h.createVacationUseCase.
}

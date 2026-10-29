package get_transfer_form

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	vacationRepo ports.VacationRepository
}

func NewHandler(vacationRepo ports.VacationRepository) *Handler {
	return &Handler{vacationRepo: vacationRepo}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	userID, err := user.ParseUserID(q.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	vacationID, err := vacation.ParseVacationID(q.VacationID)
	if err != nil {
		return nil, fmt.Errorf("parse vacation id: %w", err)
	}

	v, err := h.vacationRepo.FindByID(ctx, userID, vacationID)
	if err != nil {
		return nil, fmt.Errorf("find vacation: %w", err)
	}

	return &Result{
		UserID:       userID.String(),
		VacationID:   v.ID().String(),
		StartDate:    v.StartDate().ISO(),
		EndDate:      v.EndDate().ISO(),
		Period:       v.StartDate().Time().Format("02.01.2006") + " — " + v.EndDate().Time().Format("02.01.2006"),
		Kind:         v.Kind().Label(),
		Status:       v.Status().String(),
		Transferable: v.Status().IsActive(),
	}, nil
}

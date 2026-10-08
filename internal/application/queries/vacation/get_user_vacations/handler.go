package get_user_vacations

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type Handler struct {
	vacationRepo ports.VacationRepository
}

func NewHandler(
	vacationRepo ports.VacationRepository,
) *Handler {

	return &Handler{
		vacationRepo: vacationRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, query Query) (*Result, error) {

	userID, err := user.ParseUserID(query.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse userID: %w", err)
	}

	vacations, err := h.vacationRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find by userID: %w", err)
	}

	result := make([]Vacations, 0, len(vacations))
	for _, v := range vacations {
		result = append(result, Vacations{
			Color:     0,
			StartDate: v.StartDate().String(),
			EndDate:   v.EndDate().String(),
			Status:    v.Status().String(),
		})
	}

	return &Result{
		Vacations: result,
	}, nil
}

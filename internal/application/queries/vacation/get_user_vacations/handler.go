package get_user_vacations

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	vacationRepo  vacation.VacationRepository
}

func NewHandler(
	vacationRepo  vacation.VacationRepository,
) *Handler {

	return &Handler{
		vacationRepo: vacationRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, query Query) (*Result, error) {

	userID, err := user.ParseUserID(query.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse userID: %w")
	}
	
	vacations, err := h.vacationRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find by userID: %W")
	}

	result := make([]Vacations, 0 ,len(vacations))
	for _, v := range vacations {
		result = append(result, Vacations{
			startDate: v.StartDate().Time().GoString(), // TODO
			endDate: v.EndDate().Time().GoString(), // TODO
		})
	}

	return &Result{
		Vacations: result,
	}, nil
}
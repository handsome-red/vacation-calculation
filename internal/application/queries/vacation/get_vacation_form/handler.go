package get_vacation_form

import (
	"context"

	// "github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	repoHoliday vacation.HolidayRepository
}

func NewHandler(
	repoHoliday vacation.HolidayRepository,
	) *Handler {
	return &Handler{
		repoHoliday: repoHoliday,
	}
}

func (h *Handler) Handle(ctx context.Context, query Query) (*Result, error) {
	// userID := user.ParseUserID()
	// result, err := h.repoHoliday.ListByRange()
	return nil, nil
}
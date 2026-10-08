package get_vacation_form

import (
	"context"

	// "github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

type Handler struct {
	repoHoliday ports.HolidayRepository
}

func NewHandler(
	repoHoliday ports.HolidayRepository,
) *Handler {
	return &Handler{
		repoHoliday: repoHoliday,
	}
}

func (h *Handler) Handle(ctx context.Context, query Query) (*Result, error) {
	// userID := user.ParseUserID()
	// result, err := h.repoHoliday.ListInRange()
	return nil, nil
}

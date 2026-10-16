package new_calendar_day_form

import (
	"context"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

// TODO
type Handler struct {
	calendarDayRepo ports.CalendarDayRepository
}

func NewHandler(calendarDayRepo ports.CalendarDayRepository) *Handler {
	return &Handler{
		calendarDayRepo: calendarDayRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, _ Query) (*Result, error) {
	return nil, nil
}

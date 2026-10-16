package new_calendar_day

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	calendarDayRepo ports.CalendarDayRepository
}

func NewHandler(calendarDayRepo ports.CalendarDayRepository) *Handler {
	return &Handler{
		calendarDayRepo: calendarDayRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {

	date, err := date.ParseDate(cmd.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date %s: %w", cmd.Date, err)
	}

	dayType := date.DateType(cmd.Type)
	if !dayType.Valid() {
		return nil, fmt.Errorf("invalid day type %q", cmd.Type)
	}

	holiday, err := vacation.NewCalendarDay(date, dayType, cmd.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create holiday %s: %w", cmd.Name, err)
	}

	if err := h.calendarDayRepo.Save(ctx, holiday); err != nil {
		return nil, fmt.Errorf("failed to create holiday %s: %w", cmd.Name, err)
	}

	return &Result{
		Name: cmd.Name,
		Date: cmd.Date,
	}, nil
}

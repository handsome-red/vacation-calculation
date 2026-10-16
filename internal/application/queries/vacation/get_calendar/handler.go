package get_calendar

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/calendar"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

type Handler struct {
	calendarDayRepo ports.CalendarDayRepository
}

func NewHandler(
	calendarDayRepo ports.CalendarDayRepository,
) *Handler {
	return &Handler{
		calendarDayRepo: calendarDayRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	if q.Year < 1900 || q.Year > 2200 {
		return nil, fmt.Errorf("invalid year: %d", q.Year)
	}

	exceptions, err := h.calendarDayRepo.ListByYear(ctx, q.Year)
	if err != nil {
		return nil, fmt.Errorf("calendar days: %w", err)
	}

	exMap := make(map[date.Date]calendar.Exception, len(exceptions))
	for _, e := range exceptions {
		exMap[e.Date] = calendar.Exception{
			Date:    e.Date,
			DayType: e.DayType,
			Name:    e.Name,
		}
	}

	cal, err := calendar.BuildYear(q.Year, exMap)
	if err != nil {
		return nil, fmt.Errorf("build year: %w", err)
	}

	return &Result{Calendar: cal}, nil
}

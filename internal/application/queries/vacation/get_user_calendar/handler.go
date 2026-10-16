package get_user_calendar

import (
	"context"
	"fmt"
	"time"

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
	if q.Month < 1 || q.Month > 12 {
		return nil, fmt.Errorf("invalid month: %d", q.Month)
	}
	if q.Year < 1900 || q.Year > 2200 {
		return nil, fmt.Errorf("invalid year: %d", q.Year)
	}

	// Границы месяца: [first, last] включительно.
	first, err := date.NewDate(q.Year, time.Month(q.Month), 1)
	if err != nil {
		return nil, fmt.Errorf("month first day: %w", err)
	}
	lastDay := time.Date(q.Year, time.Month(q.Month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	last, err := date.NewDate(q.Year, time.Month(q.Month), lastDay)
	if err != nil {
		return nil, fmt.Errorf("month last day: %w", err)
	}

	calendarDays, err := h.calendarDayRepo.ListInRange(ctx, first, last)
	if err != nil {
		return nil, fmt.Errorf("calendar days: %w", err)
	}

	calendarByDate := make(map[string]date.Date, len(calendarDays))
	for _, cd := range calendarDays {
		calendarByDate[cd.Date()] = cd
	}

	days := make([]Day, 0, lastDay)
	for d := 1; d <= lastDay; d++ {
		date, err := date.NewDate(q.Year, time.Month(q.Month), d)
		if err != nil {
			return nil, fmt.Errorf("day %d: %w", d, err)
		}
		key := date.ISO()

		wd := date.Time().Weekday()
		isWeekend := wd == time.Saturday || wd == time.Sunday

		day := Day{
			Date:      date,
			IsWeekend: isWeekend,
		}

		if cd, ok := calendarByDate[key]; ok {
			switch cd.DayType() {
			case date.DateHoliday:
				day.IsHoliday = true
				day.HolidayName = cd.Name()
			case date.DateWorkSaturday:
				day.IsWorkSaturday = true
				day.WorkSaturdayName = cd.Name()
			}
		}

		days = append(days, day)
	}

	return &Result{
		Year:  q.Year,
		Month: q.Month,
		Days:  days,
	}, nil
}

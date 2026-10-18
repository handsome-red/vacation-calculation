package get_user_calendar

import (
	"context"
	"fmt"
	"time"

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

	start := date.New(q.Year, time.January, 1)
	end := date.New(q.Year,  time.December, 31)

	exceptionsDays, err := h.calendarDayRepo.ListInRange(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("calendar days: %w", err)
	}

	calendarByDate := make(map[string]calendar.Exception, len(exceptionsDays))
	for _, cd := range exceptionsDays {
		calendarByDate[cd.Date().ISO()] = cd
	}

	lastDay := time.Date(q.Year, time.Month(q.Month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
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
			case calendar.DayTypeHoliday:
				day.IsHoliday = true
				day.HolidayName = cd.Name()
			case calendar.DayTypeWorking:
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

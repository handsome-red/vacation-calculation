package new_calendar_day_form

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/calendar"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

type Handler struct {
	calendarDayRepo ports.CalendarDayRepository
}

func NewHandler(calendarDayRepo ports.CalendarDayRepository) *Handler {
	return &Handler{
		calendarDayRepo: calendarDayRepo,
	}
}

var weekdays = [...]string{"вс", "пн", "вт", "ср", "чт", "пт", "сб"}

// Handle — особые дни производственного календаря за год.
func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	days, err := h.calendarDayRepo.ListByYear(ctx, q.Year)
	if err != nil {
		return nil, fmt.Errorf("calendar days: %w", err)
	}

	out := make([]Day, 0, len(days))
	for _, d := range days {
		t := d.Date().Time()
		out = append(out, Day{
			Date:      d.Date().ISO(),
			DateLabel: t.Format("02.01.2006") + ", " + weekdays[t.Weekday()],
			Type:      string(d.DayType()),
			TypeLabel: typeLabel(d.DayType()),
			Name:      d.Name(),
		})
	}
	return &Result{Year: q.Year, Days: out}, nil
}

func typeLabel(t calendar.DayType) string {
	switch t {
	case calendar.DayTypeHoliday:
		return "Праздничный (в отпуск не входит)"
	case calendar.DayTypeWeekend:
		return "Перенесённый выходной"
	case calendar.DayTypeWorking:
		return "Рабочий день"
	}
	return string(t)
}

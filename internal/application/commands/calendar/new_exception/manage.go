package new_calendar_day

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/calendar"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

// Delete убирает особый день из производственного календаря.
func (h *Handler) Delete(ctx context.Context, day string) error {
	d, err := date.ParseDate(day)
	if err != nil {
		return fmt.Errorf("failed to parse date %s: %w", day, err)
	}
	return h.calendarDayRepo.Delete(ctx, d)
}

// FillFixedHolidays добавляет постоянные праздники года, которых ещё нет в календаре;
// уже внесённые администратором дни не меняются. Возвращает число добавленных.
func (h *Handler) FillFixedHolidays(ctx context.Context, year int) (int, error) {
	existing, err := h.calendarDayRepo.ListByYear(ctx, year)
	if err != nil {
		return 0, fmt.Errorf("calendar days: %w", err)
	}
	have := make(map[date.Date]bool, len(existing))
	for _, e := range existing {
		have[e.Date()] = true
	}

	added := 0
	for _, holiday := range calendar.FixedHolidays(year) {
		if have[holiday.Date()] {
			continue
		}
		if err := h.calendarDayRepo.Save(ctx, holiday); err != nil {
			return added, err
		}
		added++
	}
	return added, nil
}

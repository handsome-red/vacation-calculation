package get_user_calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/calendar"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	calendarDayRepo ports.CalendarDayRepository
	vacationRepo    ports.VacationRepository
}

func NewHandler(
	calendarDayRepo ports.CalendarDayRepository,
	vacationRepo ports.VacationRepository,
) *Handler {
	return &Handler{
		calendarDayRepo: calendarDayRepo,
		vacationRepo:    vacationRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	if q.Year < 1900 || q.Year > 2200 {
		return nil, fmt.Errorf("invalid year: %d", q.Year)
	}

	start := date.New(q.Year, time.January, 1)
	end := date.New(q.Year, time.December, 31)

	exceptionsDays, err := h.calendarDayRepo.ListInRange(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("calendar days: %w", err)
	}

	userID, err := user.ParseUserID(q.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	vacations, err := h.vacationRepo.FindByUserIDInRange(ctx, userID, date.New(q.Year, time.January, 1), date.New(q.Year, time.December, 31))
	if err != nil {
		return nil, fmt.Errorf("vacations: %w", err)
	}

	calendarByDate := make(map[string]calendar.Exception, len(exceptionsDays))
	for _, cd := range exceptionsDays {
		calendarByDate[cd.Date().ISO()] = cd
	}

	daysInYear := 365
	if isLeapYear(q.Year) {
		daysInYear = 366
	}

	days := make([]Day, 0, daysInYear)

	cur := start
	for i := 0; i < daysInYear; i++ {
		key := cur.ISO()

		wd := cur.Time().Weekday()
		isWeekend := wd == time.Saturday || wd == time.Sunday

		day := Day{
			Date:      cur,
			IsWeekend: isWeekend,
		}

		if cd, ok := calendarByDate[key]; ok {
			switch cd.DayType() {
			case calendar.DayTypeHoliday:
				day.IsHoliday = true
				day.HolidayName = cd.Name()
			case calendar.DayTypeWeekend:
				day.IsWeekend = true
				day.HolidayName = cd.Name()
			case calendar.DayTypeWorking:
				day.IsWorkSaturday = true
				day.WorkSaturdayName = cd.Name()
			}
		}

		days = append(days, day)
		cur = cur.AddDay()
	}

	activeVacations := make([]*vacation.Vacation, 0, len(vacations))
	for _, v := range vacations {
		if v == nil || !v.Status().IsActive() {
			continue
		}
		activeVacations = append(activeVacations, v)
	}

	return &Result{
		Year:      q.Year,
		Days:      days,
		Vacations: activeVacations,
	}, nil
}

func isLeapYear(y int) bool {
	return (y%4 == 0 && y%100 != 0) || y%400 == 0
}

package get_calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	holidayRepo  ports.HolidayRepository
	vacationRepo ports.VacationRepository
	shiftRepo    ports.ShiftRepository
}

func NewHandler(
	holidayRepo ports.HolidayRepository,
	vacationRepo ports.VacationRepository,
	shiftRepo ports.ShiftRepository,
) *Handler {
	return &Handler{
		holidayRepo:  holidayRepo,
		vacationRepo: vacationRepo,
		shiftRepo:    shiftRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	if q.Month < 1 || q.Month > 12 {
		return nil, fmt.Errorf("invalid month: %d", q.Month)
	}
	if q.Year < 1900 || q.Year > 2200 {
		return nil, fmt.Errorf("invalid year: %d", q.Year)
	}

	userID, err := user.ParseUserID(q.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Границы месяца: [first, last] включительно.
	first, err := vacation.NewDate(q.Year, time.Month(q.Month), 1)
	if err != nil {
		return nil, fmt.Errorf("month first day: %w", err)
	}
	lastDay := time.Date(q.Year, time.Month(q.Month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	last, err := vacation.NewDate(q.Year, time.Month(q.Month), lastDay)
	if err != nil {
		return nil, fmt.Errorf("month last day: %w", err)
	}

	// Три запроса к БД.
	holidays, err := h.holidayRepo.ListInRange(ctx, first, last)
	if err != nil {
		return nil, fmt.Errorf("holidays: %w", err)
	}

	vacations, err := h.vacationRepo.FindByUserIDInRange(ctx, userID, first, last)
	if err != nil {
		return nil, fmt.Errorf("vacations: %w", err)
	}

	shifts, err := h.shiftRepo.ListByUserInRange(ctx, userID, first.Time(), last.Time())
	if err != nil {
		return nil, fmt.Errorf("shifts: %w", err)
	}

	// Индексы по дате.
	holidayByDate := make(map[string]vacation.Holiday, len(holidays))
	for _, hd := range holidays {
		holidayByDate[hd.Date().ISO()] = hd
	}

	eventsByDate := make(map[string][]Event)

	for _, v := range vacations {
		if v.Status() == vacation.StatusCancelled || v.Status() == vacation.StatusRejected {
			continue
		}
		ev := Event{
			Kind:    EventVacation,
			SubKind: v.Status().String(),
			Title:   vacationTitle(v),
			From:    v.StartDate(),
			To:      v.EndDate(),
		}
		addEventToRange(eventsByDate, ev, first, last)
	}

	for _, s := range shifts {
		ev := Event{
			Kind:    EventShift,
			SubKind: string(s.Kind),
			Title:   shiftTitle(s),
			From:    dateOf(s.From),
			To:      dateOf(s.To),
		}
		addEventToRange(eventsByDate, ev, first, last)
	}

	// Собираем дни месяца.
	days := make([]Day, 0, lastDay)
	for d := 1; d <= lastDay; d++ {
		date, _ := vacation.NewDate(q.Year, time.Month(q.Month), d)
		key := date.ISO()

		wd := date.Time().Weekday()
		isWeekend := wd == time.Saturday || wd == time.Sunday

		day := Day{
			Date:      date,
			IsWeekend: isWeekend,
			Events:    eventsByDate[key],
		}
		if hd, ok := holidayByDate[key]; ok {
			day.IsHoliday = true
			day.HolidayName = hd.Name()
		}
		days = append(days, day)
	}

	return &Result{
		Year:  q.Year,
		Month: q.Month,
		Days:  days,
	}, nil
}

// addEventToRange добавляет событие во все дни месяца,
// которые попадают в его диапазон [ev.From, ev.To].
func addEventToRange(dst map[string][]Event, ev Event, monthFirst, monthLast vacation.Date) {
	start := ev.From
	if start.Before(monthFirst) {
		start = monthFirst
	}
	end := ev.To
	if end.Time().After(monthLast.Time()) {
		end = monthLast
	}
	if end.Before(start) {
		return
	}

	for d := start; !d.Time().After(end.Time()); d = nextDay(d) {
		key := d.ISO()
		dst[key] = append(dst[key], ev)
	}
}

func nextDay(d vacation.Date) vacation.Date {
	t := d.Time().AddDate(0, 0, 1)
	nd, _ := vacation.NewDate(t.Year(), t.Month(), t.Day())
	return nd
}

func dateOf(t time.Time) vacation.Date {
	d, _ := vacation.NewDate(t.Year(), t.Month(), t.Day())
	return d
}

func vacationTitle(v *vacation.Vacation) string {
	days := int(v.EndDate().Time().Sub(v.StartDate().Time()).Hours()/24) + 1
	switch v.Status() {
	case vacation.StatusDraft:
		return fmt.Sprintf("Черновик отпуска, %d дн.", days)
	default:
		return fmt.Sprintf("Отпуск, %d дн.", days)
	}
}

func shiftTitle(s vacation.Shift) string {
	days := int(s.To.Sub(s.From).Hours()/24) + 1
	switch s.Kind {
	case vacation.ShiftKindUnpaid:
		return fmt.Sprintf("Отпуск за свой счёт, %d дн.", days)
	case vacation.ShiftKindParentalLeave:
		return fmt.Sprintf("Уход за ребёнком, %d дн.", days)
	case vacation.ShiftKindAbsenteeism:
		return fmt.Sprintf("Прогул, %d дн.", days)
	default:
		return fmt.Sprintf("Сдвиг, %d дн.", days)
	}
}

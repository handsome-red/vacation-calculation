package get_vacation_form

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	userRepo        ports.UserRepository
	vacationRepo    ports.VacationRepository
	shiftRepo       ports.ShiftRepository
	calendarDayRepo ports.CalendarDayRepository
	workYearCalc    vacation.WorkYearCalculator
}

func NewHandler(
	userRepo ports.UserRepository,
	vacationRepo ports.VacationRepository,
	shiftRepo ports.ShiftRepository,
	calendarDayRepo ports.CalendarDayRepository,
	workYearCalc vacation.WorkYearCalculator,
) *Handler {
	return &Handler{
		userRepo:        userRepo,
		vacationRepo:    vacationRepo,
		shiftRepo:       shiftRepo,
		calendarDayRepo: calendarDayRepo,
		workYearCalc:    workYearCalc,
	}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	userID, err := user.ParseUserID(q.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	u, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	shifts, err := h.shiftRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list shifts: %w", err)
	}

	now := time.Now()
	hiredAt := u.HiredAt().Time()

	stats, err := h.workYearCalc.FindYearStat(
		hiredAt, now, u.IsIrregular(), shifts, h.workYearCalc.SeniorityAtFromHired(hiredAt, u.ExperienceDays()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate year stat: %w", err)
	}

	vacations, err := h.vacationRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list vacations: %w", err)
	}

	exceptions, err := h.calendarDayRepo.ListInRange(ctx, vacation.DateOf(hiredAt), vacation.HolidaysRangeEnd(vacations, vacation.DateOf(now)))
	if err != nil {
		return nil, fmt.Errorf("failed to list calendar days: %w", err)
	}
	holidays := vacation.HolidaysFrom(exceptions)

	result := &Result{
		UserID:        userID.String(),
		HasActiveYear: len(stats) > 0,
	}

	// YearStats — по каждому рабочему году; Blocked — списано из квоты этого года.
	result.YearStats = make([]YearStat, 0, len(stats))
	for _, s := range stats {
		used := vacation.UsedDaysInYear(vacations, s, holidays)
		result.YearStats = append(result.YearStats, YearStat{
			Year:      formatYearRange(s),
			Start:     vacation.DateOf(s.From).ISO(),
			Base:      Quota{Total: s.Base, Blocked: used.Base},
			Seniority: Quota{Total: s.Seniority, Blocked: used.Seniority},
			Irregular: Quota{Total: s.Irregular, Blocked: used.Irregular},
		})
	}

	if len(stats) > 0 {
		last := stats[len(stats)-1]
		used := vacation.UsedDaysInYear(vacations, last, holidays)
		result.Total = last.Base + last.Seniority + last.Irregular
		result.Used = used.Base + used.Irregular + used.Seniority

		// Доступно на сегодня с учётом остатков прошлых лет.
		balances, err := vacation.BuildBalances(stats, vacations, holidays, vacation.DateOf(now), h.workYearCalc, true)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate balances: %w", err)
		}
		for _, b := range balances {
			result.Available += b.Total()
		}
		result.Warnings = vacation.CarryoverMessages(balances, vacations, stats, vacation.DateOf(now))
	}

	return result, nil
}

// formatYearRange — "01.01.24-01.01.25" из YearStat.From/To.
func formatYearRange(s vacation.YearStat) string {
	return fmt.Sprintf("%s-%s",
		s.From.Format("02.01.06"),
		s.To.Format("02.01.06"),
	)
}

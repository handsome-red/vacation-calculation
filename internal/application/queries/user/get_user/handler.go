package get_user

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
	logger          ports.Logger
}

func NewHandler(
	userRepo ports.UserRepository,
	vacationRepo ports.VacationRepository,
	shiftRepo ports.ShiftRepository,
	calendarDayRepo ports.CalendarDayRepository,
	workYearCalc vacation.WorkYearCalculator,
	logger ports.Logger,
) *Handler {
	return &Handler{
		userRepo:        userRepo,
		vacationRepo:    vacationRepo,
		shiftRepo:       shiftRepo,
		calendarDayRepo: calendarDayRepo,
		workYearCalc:    workYearCalc,
		logger:          logger,
	}
}

func (h *Handler) Handle(ctx context.Context, query Query) (*Result, error) {
	if query.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Парсим ID
	userID, err := user.ParseUserID(query.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Ищем пользователя
	u, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		h.logger.Error(ctx, "failed to find user", "user_id", query.UserID, "error", err)
		return nil, fmt.Errorf("user not found: %w", err)
	}

	shifts, err := h.shiftRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("shifts list by user: %w", err)
	}

	now := time.Now()
	hiredAt := u.HiredAt().Time()

	stats, err := h.workYearCalc.FindYearStat(hiredAt, now, u.IsIrregular(), shifts, h.workYearCalc.SeniorityAtFromHired(hiredAt, u.ExperienceDays()))
	if err != nil {
		return nil, fmt.Errorf("find year stat: %w", err)
	}

	vacations, err := h.vacationRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("vacation by user id: %w", err)
	}

	exceptions, err := h.calendarDayRepo.ListInRange(ctx, vacation.DateOf(hiredAt), vacation.HolidaysRangeEnd(vacations, vacation.DateOf(now)))
	if err != nil {
		return nil, fmt.Errorf("calendar days: %w", err)
	}
	holidays := vacation.HolidaysFrom(exceptions)

	var current vacation.CurrentYearInfo
	carryover := 0
	var warnings []string
	if len(stats) > 0 {
		last := stats[len(stats)-1]
		used := vacation.UsedDaysInYear(vacations, last, holidays)
		current, err = h.workYearCalc.CurrentYear(stats, now, used.Base+used.Irregular+used.Seniority)
		if err != nil {
			return nil, err
		}

		balances, err := vacation.BuildBalances(stats, vacations, holidays, vacation.DateOf(now), h.workYearCalc, true)
		if err != nil {
			return nil, fmt.Errorf("balances: %w", err)
		}
		for _, b := range balances {
			if !b.Year.From.Equal(last.From) {
				carryover += b.Total()
			}
		}
		warnings = vacation.CarryoverMessages(balances, vacations, stats, vacation.DateOf(now))
	}

	return &Result{
		User:      u,
		YearStats: stats,
		Current:   current,
		Carryover: carryover,
		Warnings:  warnings,
		Now:       now,
	}, nil
}

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
	userRepo     ports.UserRepository
	vacationRepo ports.VacationRepository
	shiftRepo    ports.ShiftRepository
	workYearCalc vacation.WorkYearCalculator
}

func NewHandler(
	userRepo ports.UserRepository,
	vacationRepo ports.VacationRepository,
	shiftRepo ports.ShiftRepository,
	workYearCalc vacation.WorkYearCalculator,
) *Handler {
	return &Handler{
		userRepo:     userRepo,
		vacationRepo: vacationRepo,
		shiftRepo:    shiftRepo,
		workYearCalc: workYearCalc,
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

	stats, err := h.workYearCalc.FindYearStat(hiredAt, now, u.IsIrregular(), shifts, h.workYearCalc.SeniorityAtFromHired(hiredAt))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate year stat: %w", err)
	}

	result := &Result{
		UserID:        userID.String(),
		HasActiveYear: len(stats) > 0,
	}

	if len(stats) > 0 {
		last := stats[len(stats)-1]
		vacations, err := h.vacationRepo.FindByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to list vacations: %w", err)
		}

		used := vacation.UsedDaysInYear(vacations, last)
		totalEntitlement := last.Base + last.Seniority + last.Irregular
		available := totalEntitlement - used.Base

		result.Total = totalEntitlement
		result.Used = used.Base
		result.Available = available
	}

	return result, nil
}

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

	stats, err := h.workYearCalc.FindYearStat(
		hiredAt, now, u.IsIrregular(), shifts, h.workYearCalc.SeniorityAtFromHired(hiredAt),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate year stat: %w", err)
	}

	vacations, err := h.vacationRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list vacations: %w", err)
	}

	result := &Result{
		UserID:        userID.String(),
		HasActiveYear: len(stats) > 0,
	}

	// YearStats — по каждому рабочему году.
	result.YearStats = make([]YearStat, 0, len(stats))
	for _, s := range stats {
		used := vacation.UsedDaysInYear(vacations, s)
		result.YearStats = append(result.YearStats, YearStat{
			Year:      formatYearRange(s),
			Base:      Quota{Total: s.Base, Blocked: used.Base},
			Seniority: Quota{Total: s.Seniority, Blocked: used.Seniority},
			Irregular: Quota{Total: s.Irregular, Blocked: used.Irregular},
		})
	}

	// Total/Used/Available — по последнему году (как было).
	if len(stats) > 0 {
		last := stats[len(stats)-1]
		used := vacation.UsedDaysInYear(vacations, last)
		totalEntitlement := last.Base + last.Seniority + last.Irregular
		result.Total = totalEntitlement
		result.Used = used.Base
		result.Available = totalEntitlement - used.Base
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

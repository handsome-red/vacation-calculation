package get_vacation_form

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/leavecalc"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	userRepo ports.UserRepository
	leave    *leavecalc.Service
}

func NewHandler(userRepo ports.UserRepository, leave *leavecalc.Service) *Handler {
	return &Handler{userRepo: userRepo, leave: leave}
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

	today := vacation.DateOf(time.Now())
	snap, err := h.leave.Load(ctx, u, today, "")
	if err != nil {
		return nil, err
	}

	result := &Result{
		UserID:        userID.String(),
		HasActiveYear: len(snap.Stats) > 0,
	}

	// YearStats — по каждому рабочему году; Blocked — списано из квоты этого года.
	result.YearStats = make([]YearStat, 0, len(snap.Stats))
	for _, s := range snap.Stats {
		used := vacation.UsedDaysInYear(snap.Vacations, s, snap.Holidays)
		result.YearStats = append(result.YearStats, YearStat{
			Year:      formatYearRange(s),
			Start:     vacation.DateOf(s.From).ISO(),
			Base:      Quota{Total: s.Base, Blocked: used.Base},
			Seniority: Quota{Total: s.Seniority, Blocked: used.Seniority},
			Irregular: Quota{Total: s.Irregular, Blocked: used.Irregular},
		})
	}

	if len(snap.Stats) > 0 {
		last := snap.Stats[len(snap.Stats)-1]
		used := vacation.UsedDaysInYear(snap.Vacations, last, snap.Holidays)
		result.Total = last.Base + last.Seniority + last.Irregular
		result.Used = used.Base + used.Irregular + used.Seniority

		// Доступно на сегодня с учётом остатков прошлых лет.
		balances, err := snap.Balances(today, true)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate balances: %w", err)
		}
		for _, b := range balances {
			result.Available += b.Total()
		}
		if result.Warnings, err = snap.Warnings(today); err != nil {
			return nil, err
		}
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

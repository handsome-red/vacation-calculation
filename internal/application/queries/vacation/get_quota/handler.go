package get_quota

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type Handler struct {
	calendarDayRepo ports.CalendarDayRepository
	vacationRepo    ports.VacationRepository
	userRepo        ports.UserRepository
}

func NewHandler(
	calendarDayRepo ports.CalendarDayRepository,
	vacationRepo ports.VacationRepository,
	userRepo ports.UserRepository,
) *Handler {
	return &Handler{
		calendarDayRepo: calendarDayRepo,
		vacationRepo:    vacationRepo,
		userRepo:        userRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	uid, err := user.ParseUserID(q.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}

	u, err := h.userRepo.FindByID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	vacations, err := h.vacationRepo.FindByUserID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("vacations: %w", err)
	}

	blockedBase := 0
	for _, v := range vacations {
		blockedBase += workingDaysInRange(v.StartDate(), v.EndDate())
	}

	baseTotal := 30
	seniorityTotal := 0
	irregularTotal := 0

	return &Result{
		Years: []YearQuota{},
	}, nil
}

func workingDaysInRange(from, to date.Date) int {
	count := 0
	for cur := from; !cur.After(to); cur = cur.AddDay() {
		wd := cur.Time().Weekday()
		if wd != time.Saturday && wd != time.Sunday {
			count++
		}
	}

	return count
}

package approve_schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Command struct {
	Year         int
	DistrictCode string // пусто — все территориальные отделы
}

type Result struct {
	Approved int // утверждено записей отпусков
}

type Handler struct {
	userRepo     ports.UserRepository
	vacationRepo ports.VacationRepository
}

func NewHandler(userRepo ports.UserRepository, vacationRepo ports.VacationRepository) *Handler {
	return &Handler{userRepo: userRepo, vacationRepo: vacationRepo}
}

// Handle утверждает график: все черновики ежегодного отпуска, начинающиеся
// в году, у работающих сотрудников территориального отдела. Нарушения графика
// показываются администратору на странице, но утверждение не блокируют.
func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	users, _, err := h.userRepo.FindUsers(ctx, ports.Query{
		Page: 1, Size: 100000,
		Filters: ports.UserFilter{Status: string(user.StatusActive), DistrictCode: cmd.DistrictCode},
	})
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}

	now := time.Now()
	res := &Result{}
	for _, u := range users {
		vacations, err := h.vacationRepo.FindByUserID(ctx, u.ID())
		if err != nil {
			return res, fmt.Errorf("vacations: %w", err)
		}
		for _, v := range vacations {
			if v.Status() != vacation.StatusDraft || !v.Kind().IsAnnual() || v.StartDate().Year() != cmd.Year {
				continue
			}
			if err := v.Approve(now); err != nil {
				return res, err
			}
			if err := h.vacationRepo.Save(ctx, v); err != nil {
				return res, fmt.Errorf("save vacation: %w", err)
			}
			res.Approved++
		}
	}
	return res, nil
}

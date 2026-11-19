// Package freeze_work_years — сохранение закончившихся рабочих лет всех сотрудников.
package freeze_work_years

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/leavecalc"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Result struct {
	Users  int
	Failed int
}

type Handler struct {
	userRepo ports.UserRepository
	leave    *leavecalc.Service
	logger   ports.Logger
}

func NewHandler(userRepo ports.UserRepository, leave *leavecalc.Service, logger ports.Logger) *Handler {
	return &Handler{userRepo: userRepo, leave: leave, logger: logger}
}

// Handle сохраняет рабочие годы, закончившиеся до now, у всех сотрудников, включая
// уволенных. Ошибка по одному сотруднику не останавливает остальных.
func (h *Handler) Handle(ctx context.Context, now time.Time) (*Result, error) {
	users, _, err := h.userRepo.FindUsers(ctx, ports.Query{Page: 1, Size: 100000})
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}

	today := vacation.DateOf(now.UTC())
	res := &Result{Users: len(users)}
	for _, u := range users {
		if err := h.leave.Freeze(ctx, u, today); err != nil {
			h.logger.Error(ctx, "freeze work years", "user_id", u.ID().String(), "error", err)
			res.Failed++
		}
	}
	return res, nil
}

package get_user

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	// "github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	userRepo     ports.UserRepository
	shiftRepo    ports.ShiftRepository
	workYearCalc vacation.WorkYearCalculator
	logger       ports.Logger
}

func NewHandler(
	userRepo ports.UserRepository,
	shiftRepo ports.ShiftRepository,
	workYearCalc vacation.WorkYearCalculator,
	logger ports.Logger,
) *Handler {
	return &Handler{
		userRepo:     userRepo,
		shiftRepo:    shiftRepo,
		workYearCalc: workYearCalc,
		logger:       logger,
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

	stats, err := h.workYearCalc.FindYearStat(hiredAt, now, u.IsIrregular(), shifts, h.workYearCalc.SeniorityAtFromHired(hiredAt))
	if err != nil {
		return nil, fmt.Errorf("find year stat: %w", err)
	}

	return &Result{
		User:      u,
		YearStats: stats,
		Now:       now,
	}, nil
}

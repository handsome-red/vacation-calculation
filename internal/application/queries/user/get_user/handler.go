package get_user

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
	userRepo     ports.UserRepository
	leave        *leavecalc.Service
	workYearCalc vacation.WorkYearCalculator
	logger       ports.Logger
}

func NewHandler(
	userRepo ports.UserRepository,
	leave *leavecalc.Service,
	workYearCalc vacation.WorkYearCalculator,
	logger ports.Logger,
) *Handler {
	return &Handler{
		userRepo:     userRepo,
		leave:        leave,
		workYearCalc: workYearCalc,
		logger:       logger,
	}
}

func (h *Handler) Handle(ctx context.Context, query Query) (*Result, error) {
	if query.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	userID, err := user.ParseUserID(query.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	u, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		h.logger.Error(ctx, "failed to find user", "user_id", query.UserID, "error", err)
		return nil, fmt.Errorf("user not found: %w", err)
	}

	now := time.Now()
	today := vacation.DateOf(now)
	snap, err := h.leave.Load(ctx, u, today, "")
	if err != nil {
		return nil, err
	}

	var current vacation.CurrentYearInfo
	carryover := 0
	var warnings []string
	if len(snap.Stats) > 0 {
		last := snap.Stats[len(snap.Stats)-1]
		used := vacation.UsedDaysInYear(snap.Vacations, last, snap.Holidays)
		current, err = h.workYearCalc.CurrentYear(snap.Stats, now, used.Base+used.Irregular+used.Seniority)
		if err != nil {
			return nil, err
		}

		balances, err := snap.Balances(today, true)
		if err != nil {
			return nil, fmt.Errorf("balances: %w", err)
		}
		for _, b := range balances {
			if !b.Year.From.Equal(last.From) {
				carryover += b.Total()
			}
		}
		if warnings, err = snap.Warnings(today); err != nil {
			return nil, err
		}
	}

	return &Result{
		User:      u,
		YearStats: snap.Stats,
		Current:   current,
		Carryover: carryover,
		Warnings:  warnings,
		Now:       now,
	}, nil
}

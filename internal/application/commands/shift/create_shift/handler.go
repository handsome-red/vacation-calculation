package create_shift

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	shiftRepo ports.ShiftRepository
}

func NewHandler(shiftRepo ports.ShiftRepository) *Handler {
	return &Handler{
		shiftRepo: shiftRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	if cmd.UserID == "" {
		return nil, fmt.Errorf("create_shift: user ID is required")
	}
	if cmd.From.IsZero() || cmd.To.IsZero() {
		return nil, fmt.Errorf("create_shift: from/to are required")
	}
	if cmd.To.Before(cmd.From) {
		return nil, fmt.Errorf("create_shift: to (%s) before from (%s)",
			cmd.To.Format(vacation.DateFormat), cmd.From.Format(vacation.DateFormat))
	}

	userID, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("create_shift: invalid user ID: %w", err)
	}

	kind := vacation.ShiftKind(cmd.Kind)
	if _, ok := vacation.ShiftKindName(kind); !ok {
		return nil, fmt.Errorf("create_shift: unknown shift kind: %q", cmd.Kind)
	}

	shift := vacation.Shift{
		Kind: kind,
		From: cmd.From,
		To:   cmd.To,
	}

	if err := shift.Validate(); err != nil {
		return nil, fmt.Errorf("create_shift: %w", err)
	}

	if err := h.shiftRepo.Save(ctx, userID, shift); err != nil {
		return nil, fmt.Errorf("create_shift: save shift: %w", err)
	}

	return &Result{}, nil
}

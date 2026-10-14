package approve_vacation

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	vacationRepo ports.VacationRepository
	userRepo     ports.UserRepository
}

func NewHandler(
	vacationRepo ports.VacationRepository,
	userRepo ports.UserRepository,
) *Handler {
	return &Handler{
		vacationRepo: vacationRepo,
		userRepo:     userRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	userID, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	vacationID, err := vacation.ParseVacationID(cmd.VacationID)
	if err != nil {
		return nil, fmt.Errorf("invalid vacation ID: %w", err)
	}

	// user, err := h.userRepo.FindByID(ctx, userID)
	// if err != nil {
	// 	return nil, fmt.Errorf("find by user ID: %w", err)
	// }

	// v, err := h.vacationRepo.FindByUserIDInRange(ctx, userID, user.)
	v, err := h.vacationRepo.FindByID(ctx, userID, vacationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find vacation: %w", err)
	}

	now := time.Now()
	if err := v.Approve(now); err != nil {
		return nil, fmt.Errorf("approve failed: %w", err)
	}

	if err := h.vacationRepo.Save(ctx, v); err != nil {
		return nil, fmt.Errorf("save approved vacation: %w", err)
	}

	return &Result{
		UserID:     cmd.UserID,
		VacationID: cmd.VacationID,
	}, nil
}

package change_status

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
}

func NewHandler(vacationRepo ports.VacationRepository) *Handler {
	return &Handler{vacationRepo: vacationRepo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	userID, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	vacationID, err := vacation.ParseVacationID(cmd.VacationID)
	if err != nil {
		return fmt.Errorf("invalid vacation ID: %w", err)
	}
	status, err := vacation.NewVacationStatus(cmd.Status)
	if err != nil {
		return fmt.Errorf("invalid vacation status: %w", err)
	}

	v, err := h.vacationRepo.FindByID(ctx, userID, vacationID)
	if err != nil {
		return fmt.Errorf("failed to find vacation: %w", err)
	}
	if err := v.ChangeStatus(status, time.Now()); err != nil {
		return fmt.Errorf("change vacation status: %w", err)
	}
	if err := h.vacationRepo.Save(ctx, v); err != nil {
		return fmt.Errorf("save vacation status: %w", err)
	}
	return nil
}

package create_vacation

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

// TODO: Убрать прямой импорт пакета user
type Handler struct {
	vacationRepo vacation.VacationRepository
	userRepo     user.UserRepository
}

func NewHandler(
	vacationRepo vacation.VacationRepository,
) *Handler {
	return &Handler{
		vacationRepo: vacationRepo,
	}
}

func (h *Handler) CreateVacation(ctx context.Context, cmd Command) (*Result, error) {

	status, err := vacation.NewVacationStatus(cmd.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to create vacation status: %w", err)
	}

	userID, err := user.NewUserID(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user id: %w", err)
	}

	startDate, err := vacation.NewDate()
	if err != nil {
		return nil, fmt.Errorf("failed to create start date: %w", err)
	}

	endDate, err := vacation.NewDate()
	if err != nil {
		return nil, fmt.Errorf("failed to create end date: %w", err)
	}

	vacationID := vacation.NewVacationID()

	now := time.Now()

	newVacation, err := vacation.NewVacation(
		vacationID,
		status,
		userID,
		startDate,
		endDate,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create vacation: %w", err)
	}

	if err := h.vacationRepo.Save(ctx, newVacation); err != nil {
		return nil, fmt.Errorf("failed to save vacation: %w", err)
	}

	return &Result{
		VacationID: vacationID.String(),
		Error:      err,
	}, nil
}

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

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {

	userID, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user id: %w", err)
	}

	startDate, err := vacation.ParseDate(cmd.StartDate)
	if err != nil {
		return nil, fmt.Errorf("failed to create start date: %w", err)
	}

	endDate, err := vacation.ParseDate(cmd.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to create end date: %w", err)
	}

	vacationID := vacation.NewVacationID()

	now := time.Now()

	newVacation, err := vacation.NewVacation(
		vacationID,
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

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
	shiftRepo    vacation.ShiftRepository
	workYearCalc vacation.WorkYearCalculator
	logger       ports.Logger
}

func NewHandler(
	userRepo ports.UserRepository,
	shiftRepo vacation.ShiftRepository,
	// workYearCalc vacation.WorkYearCalculator,
	logger ports.Logger,
) *Handler {
	return &Handler{
		userRepo:  userRepo,
		shiftRepo: shiftRepo,
		// workYearCalc: workYearCalc,
		logger: logger,
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

	// shifts, err := h.shiftRepo.ListByUser(ctx, userID)
	// if err != nil {
	// 	return nil, fmt.Errorf("shifts list by user: %w", err)
	// }

	now := time.Now()

	// _, _ = h.workYearCalc.Calculate(u.HiredAt().Time(), now, shifts)

	// vacationStats := make([]VacationStat, 0, len(0))

	return &Result{
		ID:              u.ID().String(),
		Status:          u.Status().String(),
		LastName:        u.LastName(),
		FirstName:       u.FirstName(),
		MiddleName:      u.MiddleName(),
		BirthDate:       u.BirthDate().HumanRead(),
		Position:        u.Position().Title(),
		HiredAt:         u.HiredAt().String(),
		Department:      u.Department().Title,
		District:        u.District().Title(),
		WorkdayDuration: u.WorkdayDuration().Int(),
		Email:           u.Email().String(),
		IsInvalid:       u.IsInvalid(),
		Experience:      u.Experience().String(),

		//

		Initials: u.Initials(),
		Today:    now.UTC().Format("02.01.2006"),
		WorkYear: u.HiredAt().WorkYear(now).String(),
		Supposed: u.Supposed().HumanRead(),
		Earned:   0.0, // TODO

		// Vacations: vacationStats,
	}, nil
}

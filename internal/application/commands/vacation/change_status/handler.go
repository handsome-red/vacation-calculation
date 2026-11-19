package change_status

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
	vacationRepo ports.VacationRepository
	leave        *leavecalc.Service
}

func NewHandler(vacationRepo ports.VacationRepository, leave *leavecalc.Service) *Handler {
	return &Handler{vacationRepo: vacationRepo, leave: leave}
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
	save := func() error { return h.vacationRepo.Save(ctx, v) }
	// Смена статуса административного отпуска может сдвинуть рабочие годы.
	if v.Kind() == vacation.KindUnpaid {
		var adding []vacation.Shift
		if v.Status() == vacation.StatusApproved {
			adding = append(adding, vacation.NewShift(vacation.ShiftKindUnpaid, v.StartDate().Time(), v.EndDate().Time()))
		}
		err = h.leave.WithRebind(ctx, userID, adding, save)
	} else {
		err = save()
	}
	if err != nil {
		return fmt.Errorf("save vacation status: %w", err)
	}
	return nil
}

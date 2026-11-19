// Package delete_vacation — удаление записи отпуска из базы: для ошибочно внесённых
// записей. В отличие от отмены, запись и история её переносов не сохраняются.
package delete_vacation

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/application/leavecalc"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Command struct {
	UserID     string
	VacationID string
}

type Handler struct {
	vacationRepo ports.VacationRepository
	leave        *leavecalc.Service
}

func NewHandler(vacationRepo ports.VacationRepository, leave *leavecalc.Service) *Handler {
	return &Handler{vacationRepo: vacationRepo, leave: leave}
}

// Handle удаляет отпуск и возвращает удалённую запись (для журнала изменений).
func (h *Handler) Handle(ctx context.Context, cmd Command) (*vacation.Vacation, error) {
	userID, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user ID: %v", vacation.ErrVacationNotFound, err)
	}
	vacationID, err := vacation.ParseVacationID(cmd.VacationID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid vacation ID: %v", vacation.ErrVacationNotFound, err)
	}
	v, err := h.vacationRepo.FindByID(ctx, userID, vacationID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", vacation.ErrVacationNotFound, err)
	}

	remove := func() error { return h.vacationRepo.Delete(ctx, userID, vacationID) }
	// Утверждённый административный отпуск сдвигает рабочие годы: без него они
	// пересчитываются, как при отмене.
	if v.Kind() == vacation.KindUnpaid && v.Status() == vacation.StatusApproved {
		err = h.leave.WithRebind(ctx, userID, nil, remove)
	} else {
		err = remove()
	}
	if err != nil {
		return nil, fmt.Errorf("delete vacation: %w", err)
	}
	return v, nil
}

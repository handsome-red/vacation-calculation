package transfer_vacation

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/create_vacations"
	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	vacationRepo ports.VacationRepository
	transferRepo ports.TransferRepository
	planner      *create_vacations.Handler
}

func NewHandler(
	vacationRepo ports.VacationRepository,
	transferRepo ports.TransferRepository,
	planner *create_vacations.Handler,
) *Handler {
	return &Handler{vacationRepo: vacationRepo, transferRepo: transferRepo, planner: planner}
}

// Handle переносит отпуск на новые даты. Исходная запись остаётся в истории
// со статусом «Перенесён», новые даты оформляются новыми записями (с повторным
// распределением по квотам) с тем же статусом, а причина и даты сохраняются в Transfer.
func (h *Handler) Handle(ctx context.Context, cmd Command, caller auth.Caller) (*Result, error) {
	if !caller.IsAdmin {
		return nil, fmt.Errorf("%w: переносить отпуск может только администратор", vacation.ErrForbidden)
	}

	userID, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: некорректный идентификатор пользователя", vacation.ErrInvalidInput)
	}
	vacationID, err := vacation.ParseVacationID(cmd.VacationID)
	if err != nil {
		return nil, fmt.Errorf("%w: некорректный идентификатор отпуска", vacation.ErrInvalidInput)
	}
	newStart, err := date.ParseDate(cmd.NewStart)
	if err != nil {
		return nil, fmt.Errorf("%w: укажите дату начала", vacation.ErrInvalidInput)
	}
	newEnd, err := date.ParseDate(cmd.NewEnd)
	if err != nil {
		return nil, fmt.Errorf("%w: укажите дату окончания", vacation.ErrInvalidInput)
	}

	old, err := h.vacationRepo.FindByID(ctx, userID, vacationID)
	if err != nil {
		return nil, fmt.Errorf("%w: отпуск не найден (%v)", vacation.ErrInvalidInput, err)
	}
	if !old.Status().IsActive() {
		return nil, vacation.ErrVacationNotTransferable
	}
	if old.StartDate() == newStart && old.EndDate() == newEnd {
		return nil, fmt.Errorf("%w: новые даты совпадают с текущими", vacation.ErrInvalidInput)
	}

	now := time.Now()
	transfer, err := vacation.NewTransfer(old, newStart, newEnd, cmd.Reason, caller.UserID, now)
	if err != nil {
		return nil, err
	}

	// Распределение новых дат — как при оформлении, но без учёта переносимого отпуска.
	// Административный отпуск переносится административным: квоту не расходует.
	planType := create_vacations.TypeAnnual
	if old.Kind() == vacation.KindUnpaid {
		planType = create_vacations.TypeUnpaid
	}
	plan, err := h.planner.Preview(ctx, create_vacations.Command{
		UserID:            cmd.UserID,
		Type:              planType,
		Periods:           []create_vacations.Period{{Kind: vacation.KindBase, Start: newStart, End: newEnd}},
		ExcludeVacationID: old.ID().String(),
	}, caller)
	if err != nil {
		return nil, err
	}

	approved := old.Status() == vacation.StatusApproved
	newVacations := make([]*vacation.Vacation, 0, len(plan.Allocations))
	ids := make([]string, 0, len(plan.Allocations))
	for _, a := range plan.Allocations {
		v, err := vacation.NewVacation(vacation.NewVacationID(), userID, a.Kind, a.Start, a.End, a.WorkYearStart, now)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", vacation.ErrInvalidInput, err)
		}
		v.LinkTransferredFrom(old.ID())
		if approved {
			if err := v.Approve(now); err != nil {
				return nil, fmt.Errorf("approve transferred vacation: %w", err)
			}
		}
		newVacations = append(newVacations, v)
		ids = append(ids, v.ID().String())
	}

	if err := old.MarkTransferred(now); err != nil {
		return nil, err
	}

	// TODO: одной транзакцией. Порядок: сначала новые записи (на них ссылается
	// история), затем исходная запись и сама история переноса.
	for _, v := range newVacations {
		if err := h.vacationRepo.Save(ctx, v); err != nil {
			return nil, fmt.Errorf("save new vacation: %w", err)
		}
	}
	if err := h.vacationRepo.Save(ctx, old); err != nil {
		return nil, fmt.Errorf("save transferred vacation: %w", err)
	}
	if err := h.transferRepo.Save(ctx, transfer); err != nil {
		return nil, fmt.Errorf("save transfer: %w", err)
	}

	return &Result{TransferID: transfer.ID(), NewVacationIDs: ids, Warnings: plan.Warnings}, nil
}

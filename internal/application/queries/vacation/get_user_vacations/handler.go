package get_user_vacations

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	vacationRepo ports.VacationRepository
	transferRepo ports.TransferRepository
}

func NewHandler(
	vacationRepo ports.VacationRepository,
	transferRepo ports.TransferRepository,
) *Handler {

	return &Handler{
		vacationRepo: vacationRepo,
		transferRepo: transferRepo,
	}
}

func (h *Handler) Handle(ctx context.Context, query Query) (*Result, error) {

	userID, err := user.ParseUserID(query.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse userID: %w", err)
	}

	vacations, err := h.vacationRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find by userID: %w", err)
	}

	transfers, err := h.transferRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list transfers: %w", err)
	}
	// Перенос по ID исходного отпуска: и для самой перенесённой записи,
	// и для новых записей, созданных переносом (TransferredFrom).
	byVacation := make(map[string]vacation.Transfer, len(transfers))
	for _, t := range transfers {
		byVacation[t.VacationID().String()] = t
	}

	since := vacation.DateOf(time.Now().AddDate(-3, 0, 0))
	hidden := 0

	result := make([]Vacations, 0, len(vacations))
	for _, v := range vacations {
		if !query.All && v.EndDate().Before(since) {
			hidden++
			continue
		}
		item := Vacations{
			ID:         v.ID().String(),
			Kind:       v.Kind().Label(),
			WorkYear:   v.WorkYearStart().Time().Format("02.01.2006"),
			StartDate:  v.StartDate().Time().Format("02.01.2006"),
			EndDate:    v.EndDate().Time().Format("02.01.2006"),
			Status:     v.Status().String(),
			StatusCode: v.Status().ISO(),
			IsActive:   v.Status().IsActive(),
		}
		if t, ok := byVacation[v.ID().String()]; ok {
			item.Transfer = transferView(t, fmt.Sprintf("Перенесён на %s — %s", fmtDate(t.NewStart()), fmtDate(t.NewEnd())))
		} else if from := v.TransferredFrom(); !from.IsZero() {
			if t, ok := byVacation[from.String()]; ok {
				item.Transfer = transferView(t, fmt.Sprintf("Перенесён с %s — %s", fmtDate(t.OldStart()), fmtDate(t.OldEnd())))
			}
		}
		result = append(result, item)
	}

	return &Result{
		Vacations: result,
		Hidden:    hidden,
	}, nil
}

func transferView(t vacation.Transfer, summary string) *TransferInfo {
	return &TransferInfo{
		Summary: summary,
		Reason:  t.Reason(),
		Date:    t.CreatedAt().Local().Format("02.01.2006"),
	}
}

func fmtDate(d date.Date) string {
	return d.Time().Format("02.01.2006")
}

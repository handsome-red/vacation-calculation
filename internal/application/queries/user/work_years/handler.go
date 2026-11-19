// Package work_years — данные страницы границ рабочих лет сотрудника.
package work_years

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/leavecalc"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Year struct {
	Number int
	Start  string // YYYY-MM-DD
	End    string // YYYY-MM-DD, включительно
	Days   int
	Manual bool
	// Stored — год закончился и сохранён: его границы не пересчитываются.
	Stored bool
}

type View struct {
	UserID   string
	FullName string
	Years    []Year
	Error    string
}

type Handler struct {
	userRepo ports.UserRepository
	leave    *leavecalc.Service
}

func NewHandler(userRepo ports.UserRepository, leave *leavecalc.Service) *Handler {
	return &Handler{userRepo: userRepo, leave: leave}
}

func (h *Handler) Handle(ctx context.Context, userID string) (*View, error) {
	id, err := user.ParseUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("%w: parse user id: %v", user.ErrUserNotFound, err)
	}
	u, err := h.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	until := vacation.DateOf(time.Now().UTC())
	if hired := vacation.DateOf(u.HiredAt().Time()); hired.After(until) {
		until = hired
	}
	snap, err := h.leave.Load(ctx, u, until, "")
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	view := &View{UserID: userID, FullName: u.LastName() + " " + u.FirstName() + " " + u.MiddleName()}
	for i, y := range snap.Stats {
		last := y.To.AddDate(0, 0, -1)
		view.Years = append(view.Years, Year{
			Number: i + 1,
			Start:  y.From.Format(time.DateOnly),
			End:    last.Format(time.DateOnly),
			Days:   vacation.DaysInclusive(y.From, last),
			Manual: y.Manual,
			Stored: y.Stored,
		})
	}
	return view, nil
}

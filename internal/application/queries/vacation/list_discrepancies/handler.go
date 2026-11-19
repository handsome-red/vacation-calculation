// Package list_discrepancies — сотрудники с расхождениями в учёте отпусков: дни,
// списанные сверх нормы, и отпуска без рабочего года.
package list_discrepancies

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/leavecalc"
	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Query struct {
	// Caller — руководитель видит только сотрудников своих теротделов.
	Caller auth.Caller
}

type Row struct {
	UserID   string
	FullName string
	District string
	Messages []string
}

type Result struct {
	Rows    []Row
	Checked int
}

type Handler struct {
	userRepo ports.UserRepository
	leave    *leavecalc.Service
}

func NewHandler(userRepo ports.UserRepository, leave *leavecalc.Service) *Handler {
	return &Handler{userRepo: userRepo, leave: leave}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	users, _, err := h.userRepo.FindUsers(ctx, ports.Query{
		Page: 1, Size: 100000,
		Filters: ports.UserFilter{Status: string(user.StatusActive), DistrictCodes: q.Caller.Scope()},
	})
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}

	today := vacation.DateOf(time.Now().UTC())
	res := &Result{}
	for _, u := range users {
		if !q.Caller.Manages(u) {
			continue
		}
		res.Checked++
		snap, err := h.leave.Load(ctx, u, today, "")
		if err != nil {
			return nil, fmt.Errorf("%s: %w", u.Initials(), err)
		}
		if messages := snap.Discrepancies(); len(messages) > 0 {
			res.Rows = append(res.Rows, Row{
				UserID:   u.ID().String(),
				FullName: strings.Join(strings.Fields(u.LastName()+" "+u.FirstName()+" "+u.MiddleName()), " "),
				District: u.District().Title(),
				Messages: messages,
			})
		}
	}
	return res, nil
}

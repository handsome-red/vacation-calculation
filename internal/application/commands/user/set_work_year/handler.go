// Package set_work_year — ручное изменение границ рабочего года администратором.
package set_work_year

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/leavecalc"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

// ErrInvalidInput — данные формы некорректны; текст ошибки показывается администратору.
var ErrInvalidInput = errors.New("неверные данные")

type Command struct {
	UserID string
	Number int
	Start  string
	End    string
	// Reset — вернуть расчётный конец года.
	Reset bool
}

type Handler struct {
	userRepo     ports.UserRepository
	workYearRepo ports.WorkYearRepository
	leave        *leavecalc.Service
}

func NewHandler(
	userRepo ports.UserRepository,
	workYearRepo ports.WorkYearRepository,
	leave *leavecalc.Service,
) *Handler {
	return &Handler{userRepo: userRepo, workYearRepo: workYearRepo, leave: leave}
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, fmt.Sprintf(format, args...))
}

// Handle меняет границы начавшегося года. Начало года — это день после конца предыдущего,
// поэтому промежутков между годами нет. Следующие годы пересчитываются от новой границы,
// кроме заданных вручную концов. Отпуска и использованные дни остаются в году с тем же
// порядковым номером: иначе они отвязались бы от квот и дни снова стали бы доступны.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	id, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		return fmt.Errorf("%w: parse user id: %v", user.ErrUserNotFound, err)
	}
	u, err := h.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	today := vacation.DateOf(time.Now().UTC())
	if err := h.leave.Freeze(ctx, u, today); err != nil {
		return fmt.Errorf("freeze work years: %w", err)
	}
	snap, until, err := h.leave.LoadAll(ctx, u)
	if err != nil {
		return fmt.Errorf("load: %w", err)
	}
	started := 0
	for _, y := range snap.Stats {
		if !y.From.After(today.Time()) {
			started++
		}
	}
	if cmd.Number < 1 || cmd.Number > started {
		return invalid("рабочий год %d не найден", cmd.Number)
	}

	stored, err := h.workYearRepo.ListByUser(ctx, id)
	if err != nil {
		return fmt.Errorf("list work years: %w", err)
	}
	// Год Number начался, значит, предыдущие закончились и уже сохранены.
	if len(stored) < cmd.Number-1 {
		return fmt.Errorf("work year %d is not stored", cmd.Number-1)
	}
	before := append([]vacation.WorkYear(nil), stored[:cmd.Number-1]...)
	ends := make([]vacation.WorkYearEnd, 0)
	for _, y := range stored[min(cmd.Number, len(stored)):] {
		if y.Manual {
			ends = append(ends, vacation.WorkYearEnd{Number: y.Number, LastDay: y.End})
		}
	}

	if !cmd.Reset {
		end, changed, err := applyDates(cmd, snap.Stats[cmd.Number-1], before)
		if err != nil {
			return err
		}
		if !changed {
			return nil
		}
		ends = append(ends, vacation.WorkYearEnd{Number: cmd.Number, LastDay: end})
	}

	stats, err := snap.Years(before, ends, until)
	if err != nil {
		return fmt.Errorf("work years: %w", err)
	}
	// Конец раньше начала расчёт не применит — значит, годы задвинуты друг на друга.
	for _, e := range ends {
		if e.Number <= len(stats) && lastDayOf(stats[e.Number-1]) != e.LastDay {
			return invalid("рабочий год %d: конец раньше начала. Измените или сбросьте его границы", e.Number)
		}
	}

	// Сохраняются закончившиеся годы и годы с границами, заданными вручную.
	keep := 0
	for i, y := range stats {
		if y.Manual || !y.To.After(today.Time()) {
			keep = i + 1
		}
	}
	years := make([]vacation.WorkYear, 0, keep)
	for i, y := range stats[:keep] {
		years = append(years, leavecalc.ToWorkYear(i+1, y))
	}

	moved, opening, err := h.leave.PlanRebind(ctx, u.ID(), snap.Stats, stats, snap.Vacations)
	if errors.Is(err, leavecalc.ErrRebind) {
		return invalid("%s", err.Error())
	}
	if err != nil {
		return err
	}
	return h.workYearRepo.Apply(ctx, id, years, moved, opening)
}

// applyDates проверяет новые границы года и при изменении начала переносит конец
// предыдущего года в before; changed=false — границы не изменились.
func applyDates(cmd Command, year vacation.YearStat, before []vacation.WorkYear) (end date.Date, changed bool, err error) {
	start, err := date.ParseDate(cmd.Start)
	if err != nil {
		return date.Date{}, false, invalid("некорректная дата начала")
	}
	end, err = date.ParseDate(cmd.End)
	if err != nil {
		return date.Date{}, false, invalid("некорректная дата конца")
	}
	if end.Before(start) {
		return date.Date{}, false, invalid("конец рабочего года раньше его начала")
	}

	startChanged := start != vacation.DateOf(year.From)
	if !startChanged && end == lastDayOf(year) {
		return end, false, nil
	}
	if startChanged {
		if cmd.Number == 1 {
			return date.Date{}, false, invalid("начало первого рабочего года — дата приёма, она меняется в данных сотрудника")
		}
		prev := &before[cmd.Number-2]
		prev.End = vacation.DateOf(start.Time().AddDate(0, 0, -1))
		prev.Manual = true
		if prev.End.Before(prev.Start) {
			return date.Date{}, false, invalid("рабочий год %d: конец раньше начала. Измените или сбросьте его границы", prev.Number)
		}
	}
	return end, true, nil
}

func lastDayOf(y vacation.YearStat) date.Date {
	return vacation.DateOf(y.To.AddDate(0, 0, -1))
}

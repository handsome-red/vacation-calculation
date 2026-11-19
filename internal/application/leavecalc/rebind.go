package leavecalc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

// rebindHorizonYears — запас рабочих лет вперёд, чтобы перенести и будущие отпуска.
const rebindHorizonYears = 20

var (
	// ErrRebind — записи нельзя перенести на новые рабочие годы.
	ErrRebind = errors.New("записи нельзя перенести на новые рабочие годы")
	// ErrShiftOverlap — период пересекается с уже учтённым сдвигом.
	ErrShiftOverlap = errors.New("период пересекается с другим сдвигом")
)

// LoadAll — данные сотрудника со всеми рабочими годами в пределах запаса.
func (s *Service) LoadAll(ctx context.Context, u *user.User) (*Snapshot, date.Date, error) {
	until := vacation.DateOf(time.Now().UTC().AddDate(rebindHorizonYears, 0, 0))
	if hired := vacation.DateOf(u.HiredAt().Time()); hired.After(until) {
		until = hired
	}
	snap, err := s.Load(ctx, u, until, "")
	return snap, until, err
}

// WithRebind выполняет change, меняющий границы рабочих лет, и переносит отпуска и
// использованные дни на новые годы. Закончившиеся годы сначала сохраняются, поэтому
// change меняет только текущий и следующие. adding — новые периоды сдвига: пересечение
// с учтёнными сломало бы расчёт лет, поэтому оно отклоняется до изменения.
func (s *Service) WithRebind(ctx context.Context, userID user.UserID, adding []vacation.Shift, change func() error) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if err := s.Freeze(ctx, u, vacation.DateOf(time.Now().UTC())); err != nil {
		return fmt.Errorf("freeze work years: %w", err)
	}
	before, _, err := s.LoadAll(ctx, u)
	if err != nil {
		return err
	}
	for _, a := range adding {
		for _, existing := range before.Shifts {
			if !a.From.After(existing.To) && !existing.From.After(a.To) {
				return fmt.Errorf("%w: %s — %s", ErrShiftOverlap,
					existing.From.Format("02.01.2006"), existing.To.Format("02.01.2006"))
			}
		}
	}

	if err := change(); err != nil {
		return err
	}

	after, _, err := s.LoadAll(ctx, u)
	if err != nil {
		return err
	}
	moved, opening, err := s.PlanRebind(ctx, userID, before.Stats, after.Stats, after.Vacations)
	if err != nil {
		return err
	}
	stored, err := s.workYearRepo.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("work years: %w", err)
	}
	return s.workYearRepo.Apply(ctx, userID, stored, moved, opening)
}

// PlanRebind рассчитывает перенос записей после изменения границ лет. Отпуска с квотой и
// использованные дни остаются в году с тем же номером: иначе они отвязались бы от квот и
// дни снова стали бы доступны. Административный отпуск квоту не расходует, поэтому
// привязан к году, в который попадает его начало. Записи, не совпадающие ни с одним
// прежним годом, не меняются.
func (s *Service) PlanRebind(
	ctx context.Context,
	userID user.UserID,
	before, after []vacation.YearStat,
	vacations []*vacation.Vacation,
) ([]vacation.VacationWorkYear, []vacation.OpeningUsage, error) {
	numbers := make(map[date.Date]int, len(before))
	for i, y := range before {
		numbers[vacation.DateOf(y.From)] = i
	}
	moveTo := func(start date.Date) (date.Date, error) {
		i, ok := numbers[start]
		if !ok {
			return start, nil
		}
		if i >= len(after) {
			return start, fmt.Errorf("%w: при таких границах рабочий год %d не наступит", ErrRebind, i+1)
		}
		return vacation.DateOf(after[i].From), nil
	}

	moved := make([]vacation.VacationWorkYear, 0)
	for _, v := range vacations {
		start, err := moveTo(v.WorkYearStart())
		if err != nil {
			return nil, nil, err
		}
		if v.Kind() == vacation.KindUnpaid {
			if y, ok := yearOf(after, v.StartDate()); ok {
				start = vacation.DateOf(y.From)
			}
		}
		if start != v.WorkYearStart() {
			moved = append(moved, vacation.VacationWorkYear{ID: v.ID(), Start: start})
		}
	}

	usages, err := s.openingRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("list opening usage: %w", err)
	}
	seen := make(map[date.Date]bool, len(usages))
	for i, usage := range usages {
		start, err := moveTo(usage.WorkYearStart)
		if err != nil {
			return nil, nil, err
		}
		if seen[start] {
			return nil, nil, fmt.Errorf("%w: использованные дни двух лет попадают на начало %s", ErrRebind, start.Time().Format("02.01.2006"))
		}
		seen[start] = true
		usages[i].WorkYearStart = start
	}
	return moved, usages, nil
}

func yearOf(stats []vacation.YearStat, d date.Date) (vacation.YearStat, bool) {
	t := d.Time()
	for _, y := range stats {
		if !t.Before(y.From) && t.Before(y.To) {
			return y, true
		}
	}
	return vacation.YearStat{}, false
}

// Package leavecalc собирает всё, что нужно для расчёта отпусков сотрудника:
// рабочие годы (со сдвигами и выслугой), отпуска, праздники. Один источник для
// карточки, формы отпуска, оформления, графика, отчёта и уведомлений.
package leavecalc

import (
	"context"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Service struct {
	vacationRepo    ports.VacationRepository
	shiftRepo       ports.ShiftRepository
	calendarDayRepo ports.CalendarDayRepository
	calc            vacation.WorkYearCalculator
}

func NewService(
	vacationRepo ports.VacationRepository,
	shiftRepo ports.ShiftRepository,
	calendarDayRepo ports.CalendarDayRepository,
	calc vacation.WorkYearCalculator,
) *Service {
	return &Service{vacationRepo: vacationRepo, shiftRepo: shiftRepo, calendarDayRepo: calendarDayRepo, calc: calc}
}

// Snapshot — данные сотрудника для расчётов.
type Snapshot struct {
	User      *user.User
	Vacations []*vacation.Vacation // все записи, включая отменённые и перенесённые
	Shifts    []vacation.Shift     // исключаемые периоды + утверждённые административные отпуска
	Stats     []vacation.YearStat  // рабочие годы, начавшиеся не позже until
	Holidays  vacation.Holidays    // от приёма до until или конца последнего отпуска
	calc      vacation.WorkYearCalculator
}

// Load собирает данные сотрудника u с рабочими годами до даты until включительно.
// exclude — ID отпуска, который не учитывается (например, переносимый).
func (s *Service) Load(ctx context.Context, u *user.User, until date.Date, exclude string) (*Snapshot, error) {
	vacations, err := s.vacationRepo.FindByUserID(ctx, u.ID())
	if err != nil {
		return nil, fmt.Errorf("vacations: %w", err)
	}
	if exclude != "" {
		kept := make([]*vacation.Vacation, 0, len(vacations))
		for _, v := range vacations {
			if v.ID().String() != exclude {
				kept = append(kept, v)
			}
		}
		vacations = kept
	}

	shifts, err := s.shiftRepo.ListByUser(ctx, u.ID())
	if err != nil {
		return nil, fmt.Errorf("shifts: %w", err)
	}
	shifts = vacation.WithUnpaidLeaves(shifts, vacations)

	hiredAt := u.HiredAt().Time()
	stats, err := s.calc.FindYearStat(
		hiredAt,
		until.Time().AddDate(0, 0, 1),
		u.IsIrregular(),
		shifts,
		s.calc.SeniorityAtFromHired(hiredAt, u.ExperienceDays(), u.Position().IsCivilServant()),
	)
	if err != nil {
		return nil, fmt.Errorf("work years: %w", err)
	}

	exceptions, err := s.calendarDayRepo.ListInRange(ctx, vacation.DateOf(hiredAt), vacation.HolidaysRangeEnd(vacations, until))
	if err != nil {
		return nil, fmt.Errorf("calendar days: %w", err)
	}

	return &Snapshot{
		User:      u,
		Vacations: vacations,
		Shifts:    shifts,
		Stats:     stats,
		Holidays:  vacation.HolidaysFrom(exceptions),
		calc:      s.calc,
	}, nil
}

// Balances — остатки квот по рабочим годам на дату at (см. vacation.BuildBalances).
func (s *Snapshot) Balances(at date.Date, capEarned bool) ([]vacation.YearBalance, error) {
	return vacation.BuildBalances(s.Stats, s.Vacations, s.Holidays, at, s.calc, capEarned)
}

// YearOf — рабочий год, в который попадает дата d.
func (s *Snapshot) YearOf(d date.Date) (vacation.YearStat, bool) {
	t := d.Time()
	for _, y := range s.Stats {
		if !t.Before(y.From) && t.Before(y.To) {
			return y, true
		}
	}
	return vacation.YearStat{}, false
}

// Warnings — предупреждения по сотруднику на дату today: срок переноса
// неиспользованного отпуска и административные отпуска больше 14 дней за год.
func (s *Snapshot) Warnings(today date.Date) ([]string, error) {
	balances, err := s.Balances(today, true)
	if err != nil {
		return nil, err
	}
	return vacation.CarryoverMessages(balances, s.Vacations, s.Stats, s.Shifts, today), nil
}

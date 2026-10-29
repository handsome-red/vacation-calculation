package create_vacations

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

// TODO: Убрать прямой импорт пакета user
type Handler struct {
	vacationRepo    ports.VacationRepository
	userRepo        ports.UserRepository
	shiftRepo       ports.ShiftRepository
	calendarDayRepo ports.CalendarDayRepository
	workYearCalc    vacation.WorkYearCalculator
}

func NewHandler(
	vacationRepo ports.VacationRepository,
	userRepo ports.UserRepository,
	shiftRepo ports.ShiftRepository,
	calendarDayRepo ports.CalendarDayRepository,
	workYearCalc vacation.WorkYearCalculator,
) *Handler {
	return &Handler{
		vacationRepo:    vacationRepo,
		userRepo:        userRepo,
		shiftRepo:       shiftRepo,
		calendarDayRepo: calendarDayRepo,
		workYearCalc:    workYearCalc,
	}
}

// Handle проверяет заявку, распределяет дни по квотам рабочих годов и сохраняет
// по одной записи отпуска на каждый кусок распределения.
func (h *Handler) Handle(ctx context.Context, cmd Command, caller auth.Caller) (*Result, error) {
	userID, res, err := h.plan(ctx, cmd, caller)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	newVacations := make([]*vacation.Vacation, 0, len(res.Allocations))
	ids := make([]string, 0, len(res.Allocations))
	for _, a := range res.Allocations {
		vacationID := vacation.NewVacationID()
		v, err := vacation.NewVacation(vacationID, userID, a.Kind, a.Start, a.End, a.WorkYearStart, now)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", vacation.ErrInvalidInput, err)
		}
		newVacations = append(newVacations, v)
		ids = append(ids, vacationID.String())
	}

	// TODO: SaveBatch — одной транзакцией. Пока цикл:
	for _, v := range newVacations {
		if err := h.vacationRepo.Save(ctx, v); err != nil {
			return nil, fmt.Errorf("failed to save vacation: %w", err)
		}
	}

	res.VacationIDs = ids
	return res, nil
}

// Preview выполняет те же проверки и распределение, что и Handle, но ничего не сохраняет.
func (h *Handler) Preview(ctx context.Context, cmd Command, caller auth.Caller) (*Result, error) {
	_, res, err := h.plan(ctx, cmd, caller)
	return res, err
}

// withNotices собирает результат планирования и добавляет предупреждения
// о днях, используемых позже срока переноса (не блокируют оформление).
func withNotices(allocations []vacation.Allocation, balances []vacation.YearBalance, warnings []string) *Result {
	late := vacation.LateAllocations(allocations, balances)
	notices := make([]string, 0, len(late))
	for _, l := range late {
		notices = append(notices, l.Message())
	}
	return &Result{Allocations: allocations, Warnings: warnings, Notices: notices}
}

// plan проверяет заявку и распределяет дни. Warnings — нарушенные правила,
// которые администратор обошёл; у сотрудника любое нарушение — ошибка.
func (h *Handler) plan(ctx context.Context, cmd Command, caller auth.Caller) (user.UserID, *Result, error) {
	if err := cmd.Validate(); err != nil {
		return user.UserID{}, nil, err
	}

	userID, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		return user.UserID{}, nil, fmt.Errorf("%w: некорректный идентификатор пользователя %q", vacation.ErrInvalidInput, cmd.UserID)
	}

	if !caller.IsAdmin && caller.UserID != userID {
		return user.UserID{}, nil, fmt.Errorf("%w: нельзя оформлять отпуск другому пользователю", vacation.ErrForbidden)
	}

	u, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		return user.UserID{}, nil, fmt.Errorf("failed to find user: %w", err)
	}
	shifts, err := h.shiftRepo.ListByUser(ctx, userID)
	if err != nil {
		return user.UserID{}, nil, fmt.Errorf("failed to list shifts: %w", err)
	}

	hiredAt := u.HiredAt().Time()

	// 1. Дата начала каждого периода >= даты найма, конец не раньше начала.
	for i, p := range cmd.Periods {
		if p.End.Before(p.Start) {
			return user.UserID{}, nil, fmt.Errorf("%w: период #%d: дата окончания раньше даты начала", vacation.ErrInvalidInput, i+1)
		}
		if p.Start.Time().Before(hiredAt) {
			return user.UserID{}, nil, fmt.Errorf("%w: период #%d: начало %s раньше даты приёма на работу %s",
				vacation.ErrBeforeHireDate, i+1,
				p.Start.Time().Format("02.01.2006"),
				hiredAt.Format("02.01.2006"))
		}
	}

	domainPeriods := make([]vacation.CreationPeriod, len(cmd.Periods))
	for i, p := range cmd.Periods {
		domainPeriods[i] = vacation.CreationPeriod{Kind: p.Kind, Start: p.Start, End: p.End}
	}
	domainPeriods = vacation.MergeCreationPeriods(domainPeriods)

	// Самый ранний и самый поздний концы — для расчёта рабочих лет.
	earliest, latest := domainPeriods[0].Start, domainPeriods[0].End
	for _, p := range domainPeriods[1:] {
		if p.Start.Before(earliest) {
			earliest = p.Start
		}
		if p.End.After(latest) {
			latest = p.End
		}
	}

	stats, err := h.workYearCalc.FindYearStat(
		hiredAt,
		latest.Time().AddDate(0, 0, 1),
		u.IsIrregular(),
		shifts,
		h.workYearCalc.SeniorityAtFromHired(hiredAt, u.ExperienceDays()),
	)
	if err != nil {
		return user.UserID{}, nil, fmt.Errorf("failed to calculate work years: %w", err)
	}

	vacations, err := h.vacationRepo.FindByUserID(ctx, userID)
	if err != nil {
		return user.UserID{}, nil, fmt.Errorf("failed to list vacations: %w", err)
	}
	if cmd.ExcludeVacationID != "" {
		kept := vacations[:0]
		for _, v := range vacations {
			if v.ID().String() != cmd.ExcludeVacationID {
				kept = append(kept, v)
			}
		}
		vacations = kept
	}

	exceptions, err := h.calendarDayRepo.ListInRange(ctx, vacation.DateOf(hiredAt), vacation.HolidaysRangeEnd(vacations, latest))
	if err != nil {
		return user.UserID{}, nil, fmt.Errorf("failed to list calendar days: %w", err)
	}
	holidays := vacation.HolidaysFrom(exceptions)

	// 2. Рабочий год, в который попадает начало самого раннего периода.
	var matched *vacation.YearStat
	for i := range stats {
		t := earliest.Time()
		if !t.Before(stats[i].From) && t.Before(stats[i].To) {
			matched = &stats[i]
			break
		}
	}
	if matched == nil {
		return user.UserID{}, nil, fmt.Errorf("%w: дата начала %s не попадает ни в один рабочий год сотрудника",
			vacation.ErrNoSuitableWorkYear,
			earliest.Time().Format("02.01.2006"))
	}

	// 3. Календарные правила. Пересечение с другим отпуском запрещено всем,
	// остальные нарушения администратор может обойти — они становятся предупреждениями.
	violations, err := vacation.ValidateCreation(domainPeriods, *matched, vacations)
	if err != nil {
		return user.UserID{}, nil, err
	}
	if len(violations) > 0 && !caller.IsAdmin {
		return user.UserID{}, nil, violations[0]
	}
	warnings := make([]string, 0, len(violations)+1)
	for _, v := range violations {
		warnings = append(warnings, v.Error())
	}

	// 4. Распределение по квотам: от самого старого года с остатком.
	// Для сотрудника в текущем году доступны только заработанные дни.
	balances, err := vacation.BuildBalances(stats, vacations, holidays, earliest, h.workYearCalc, true)
	if err != nil {
		return user.UserID{}, nil, fmt.Errorf("failed to calculate balances: %w", err)
	}
	allocations, _, err := vacation.Allocate(domainPeriods, balances, holidays, false)
	if err == nil {
		return userID, withNotices(allocations, balances, warnings), nil
	}
	if !caller.IsAdmin || !errors.Is(err, vacation.ErrExceedsAnnual) {
		return user.UserID{}, nil, err
	}

	// Администратор: вся норма текущего года без ограничения заработанным,
	// нехватка списывается сверх нормы.
	warnings = append(warnings, err.Error())
	balances, err = vacation.BuildBalances(stats, vacations, holidays, earliest, h.workYearCalc, false)
	if err != nil {
		return user.UserID{}, nil, fmt.Errorf("failed to calculate balances: %w", err)
	}
	allocations, overdraft, err := vacation.Allocate(domainPeriods, balances, holidays, true)
	if err != nil {
		return user.UserID{}, nil, err
	}
	if overdraft > 0 {
		last := balances[len(balances)-1].Year.From.Format("02.01.2006")
		warnings = append(warnings, fmt.Sprintf(
			"%d дн. списаны сверх нормы основного отпуска рабочего года с %s", overdraft, last))
	}

	return userID, withNotices(allocations, balances, warnings), nil
}

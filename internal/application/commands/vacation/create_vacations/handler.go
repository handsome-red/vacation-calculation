package create_vacations

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/leavecalc"
	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

// TODO: Убрать прямой импорт пакета user
type Handler struct {
	vacationRepo ports.VacationRepository
	userRepo     ports.UserRepository
	leave        *leavecalc.Service
}

func NewHandler(
	vacationRepo ports.VacationRepository,
	userRepo ports.UserRepository,
	leave *leavecalc.Service,
) *Handler {
	return &Handler{
		vacationRepo: vacationRepo,
		userRepo:     userRepo,
		leave:        leave,
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
// Административные куски unpaid квоту не расходуют и в проверку срока не входят.
func withNotices(allocations, unpaid []vacation.Allocation, balances []vacation.YearBalance, warnings []string) *Result {
	late := vacation.LateAllocations(allocations, balances)
	notices := make([]string, 0, len(late))
	for _, l := range late {
		notices = append(notices, l.Message())
	}
	all := append(append([]vacation.Allocation{}, allocations...), unpaid...)
	sort.SliceStable(all, func(i, j int) bool { return all[i].Start.Before(all[j].Start) })
	return &Result{Allocations: all, Warnings: warnings, Notices: notices}
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

	u, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		return user.UserID{}, nil, fmt.Errorf("failed to find user: %w", err)
	}
	// Администратор и руководитель теротдела сотрудника оформляют отпуск с правами
	// администратора (нарушения правил — предупреждения), сотрудник — только себе.
	manages := caller.Manages(u)
	if !manages && caller.UserID != userID {
		return user.UserID{}, nil, fmt.Errorf("%w: нельзя оформлять отпуск другому пользователю", vacation.ErrForbidden)
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

	snap, err := h.leave.Load(ctx, u, latest, cmd.ExcludeVacationID)
	if err != nil {
		return user.UserID{}, nil, err
	}

	// 2. Рабочий год, в который попадает начало самого раннего периода.
	matched, ok := snap.YearOf(earliest)
	if !ok {
		return user.UserID{}, nil, fmt.Errorf("%w: дата начала %s не попадает ни в один рабочий год сотрудника",
			vacation.ErrNoSuitableWorkYear,
			earliest.Time().Format("02.01.2006"))
	}

	// 3. Правила. Пересечение с другим отпуском запрещено всем, остальные нарушения
	// администратор может обойти — они становятся предупреждениями.
	today := vacation.DateOf(time.Now())
	violations, err := vacation.ValidateCreation(domainPeriods, matched, snap.Vacations, snap.Holidays, today)
	if err != nil {
		return user.UserID{}, nil, err
	}
	if len(violations) > 0 && !manages {
		return user.UserID{}, nil, violations[0]
	}
	warnings := make([]string, 0, len(violations)+1)
	for _, v := range violations {
		warnings = append(warnings, v.Error())
	}

	// Административный отпуск квоту не расходует: по записи на период,
	// рабочий год — тот, в который попадает начало периода.
	unpaid := make([]vacation.Allocation, 0)
	annual := make([]vacation.CreationPeriod, 0, len(domainPeriods))
	for _, p := range domainPeriods {
		if p.Kind != vacation.KindUnpaid {
			annual = append(annual, p)
			continue
		}
		y, ok := snap.YearOf(p.Start)
		if !ok {
			y = matched
		}
		unpaid = append(unpaid, vacation.Allocation{
			Kind: vacation.KindUnpaid, WorkYearStart: vacation.DateOf(y.From),
			Start: p.Start, End: p.End, Days: vacation.DaysInclusive(p.Start.Time(), p.End.Time()),
		})
	}
	if len(annual) == 0 {
		return userID, withNotices(nil, unpaid, nil, warnings), nil
	}

	// 4. Распределение по квотам: от самого старого года с остатком.
	allocations, balances, advance, err := allocate(snap, annual, manages)
	if err != nil {
		return user.UserID{}, nil, err
	}
	if advance != "" {
		warnings = append(warnings, advance)
	}
	res := withNotices(allocations, unpaid, balances, warnings)
	for _, b := range balances {
		res.Years = append(res.Years, YearQuota{Year: b.Year, Used: vacation.UsedDaysInYear(snap.Vacations, b.Year, snap.Holidays)})
	}
	return userID, res, nil
}

// allocate распределяет дни годовых отпусков по квотам. Сотруднику в текущем году
// доступны только дни, заработанные к началу отпуска; администратору и руководителю
// (manages) — вся норма года, тогда advance — предупреждение об этом. Сверх нормы дни
// не списываются никому.
func allocate(snap *leavecalc.Snapshot, annual []vacation.CreationPeriod, manages bool) (allocations []vacation.Allocation, balances []vacation.YearBalance, advance string, err error) {
	balanceDate := annual[0].Start
	for _, p := range annual[1:] {
		if p.Start.Before(balanceDate) {
			balanceDate = p.Start
		}
	}
	balances, err = snap.Balances(balanceDate, true)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to calculate balances: %w", err)
	}
	allocations, err = vacation.Allocate(annual, balances, snap.Holidays)
	if err == nil || !manages || !errors.Is(err, vacation.ErrExceedsAnnual) {
		return allocations, balances, "", err
	}

	earnedErr := err
	balances, err = snap.Balances(balanceDate, false)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to calculate balances: %w", err)
	}
	allocations, err = vacation.Allocate(annual, balances, snap.Holidays)
	if err != nil {
		return nil, nil, "", err
	}
	return allocations, balances, earnedErr.Error() + " — списаны ещё не заработанные дни текущего рабочего года", nil
}

// limitHorizonDays — на сколько дней от начала периода ищется последний доступный день.
const limitHorizonDays = 400

// Limit — сколько дней вида Kind можно взять в новом периоде с началом Start вместе с
// уже выбранными периодами заявки cmd.Periods. LastDay — последний день такого периода
// (праздники в счёт дней не входят); Days = 0 — дней нет. Административный отпуск
// квоту не расходует: Unlimited.
type Limit struct {
	Days      int
	LastDay   date.Date
	Unlimited bool
}

// Limit считает, до какого дня можно продлить новый период: то же распределение по
// квотам, что и при оформлении, без проверки остальных правил.
func (h *Handler) Limit(ctx context.Context, cmd Command, kind vacation.Kind, start date.Date, caller auth.Caller) (*Limit, error) {
	if kind == vacation.KindUnpaid {
		return &Limit{Unlimited: true}, nil
	}
	userID, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: некорректный идентификатор пользователя %q", vacation.ErrInvalidInput, cmd.UserID)
	}
	u, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	manages := caller.Manages(u)
	if !manages && caller.UserID != userID {
		return nil, fmt.Errorf("%w: нельзя оформлять отпуск другому пользователю", vacation.ErrForbidden)
	}
	if start.Time().Before(u.HiredAt().Time()) {
		return &Limit{}, nil
	}

	annual := make([]vacation.CreationPeriod, 0, len(cmd.Periods)+1)
	horizon := vacation.DateOf(start.Time().AddDate(0, 0, limitHorizonDays))
	for _, p := range cmd.Periods {
		if p.Kind == vacation.KindUnpaid {
			continue
		}
		annual = append(annual, vacation.CreationPeriod{Kind: p.Kind, Start: p.Start, End: p.End})
		if p.End.After(horizon) {
			horizon = p.End
		}
	}
	snap, err := h.leave.Load(ctx, u, horizon, cmd.ExcludeVacationID)
	if err != nil {
		return nil, err
	}

	fits := func(end date.Date) (bool, error) {
		periods := append(append([]vacation.CreationPeriod{}, annual...), vacation.CreationPeriod{Kind: kind, Start: start, End: end})
		_, _, _, err := allocate(snap, periods, manages)
		switch {
		case err == nil:
			return true, nil
		case errors.Is(err, vacation.ErrExceedsAnnual), errors.Is(err, vacation.ErrNoSuitableWorkYear):
			return false, nil
		}
		return false, err
	}

	// Чем длиннее период, тем больше дней он списывает: ищем самый длинный, что помещается.
	lo, hi := -1, limitHorizonDays
	for lo < hi {
		mid := (lo + hi + 1) / 2
		ok, err := fits(vacation.DateOf(start.Time().AddDate(0, 0, mid)))
		if err != nil {
			return nil, err
		}
		if ok {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	if lo < 0 {
		return &Limit{}, nil
	}
	last := vacation.DateOf(start.Time().AddDate(0, 0, lo))
	return &Limit{Days: vacation.ChargeableDays(start, last, snap.Holidays), LastDay: last}, nil
}

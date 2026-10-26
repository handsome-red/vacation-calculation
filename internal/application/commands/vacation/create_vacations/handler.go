package create_vacations

import (
	"context"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

// TODO: Убрать прямой импорт пакета user
type Handler struct {
	vacationRepo ports.VacationRepository
	userRepo     ports.UserRepository
	shiftRepo    ports.ShiftRepository
	workYearCalc vacation.WorkYearCalculator
}

func NewHandler(
	vacationRepo ports.VacationRepository,
	userRepo ports.UserRepository,
	shiftRepo ports.ShiftRepository,
	workYearCalc vacation.WorkYearCalculator,
) *Handler {
	return &Handler{
		vacationRepo: vacationRepo,
		userRepo:     userRepo,
		shiftRepo:    shiftRepo,
		workYearCalc: workYearCalc,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command, caller auth.Caller) (*Result, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	userID, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: некорректный идентификатор пользователя %q", vacation.ErrInvalidInput, cmd.UserID)
	}

	if !caller.IsAdmin && caller.UserID != userID {
		return nil, fmt.Errorf("%w: нельзя оформлять отпуск другому пользователю", vacation.ErrForbidden)
	}

	u, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	shifts, err := h.shiftRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list shifts: %w", err)
	}

	hiredAt := u.HiredAt().Time()

	// 2. Дата начала каждого периода >= даты найма.
	for i, p := range cmd.Periods {
		if p.Start.Time().Before(hiredAt) {
			return nil, fmt.Errorf("%w: период #%d: начало %s раньше даты приёма на работу %s",
				vacation.ErrBeforeHireDate, i+1,
				p.Start.Time().Format("02.01.2006"),
				hiredAt.Format("02.01.2006"))
		}
	}

	// Самый ранний и самый поздний концы — для расчёта рабочего года.
	earliestStart := cmd.Periods[0].Start.Time()
	latestEnd := cmd.Periods[0].End.Time()
	for _, p := range cmd.Periods[1:] {
		if p.Start.Time().Before(earliestStart) {
			earliestStart = p.Start.Time()
		}
		if p.End.Time().After(latestEnd) {
			latestEnd = p.End.Time()
		}
	}

	stats, err := h.workYearCalc.FindYearStat(
		hiredAt,
		latestEnd.AddDate(0, 0, 1),
		u.IsIrregular(),
		shifts,
		h.workYearCalc.SeniorityAtFromHired(hiredAt),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate work years: %w", err)
	}

	vacations, err := h.vacationRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list vacations: %w", err)
	}

	// 3. Поиск рабочего года по самому раннему периоду.
	var matched *vacation.YearStat
	for i := range stats {
		if !earliestStart.Before(stats[i].From) && earliestStart.Before(stats[i].To) {
			matched = &stats[i]
			break
		}
	}
	if matched == nil {
		return nil, fmt.Errorf("%w: дата начала %s не попадает ни в один рабочий год сотрудника",
			vacation.ErrNoSuitableWorkYear,
			earliestStart.Format("02.01.2006"))
	}

	// 4. Доменная валидация по всем периодам.
	domainPeriods := make([]vacation.CreationPeriod, len(cmd.Periods))
	for i, p := range cmd.Periods {
		domainPeriods[i] = vacation.CreationPeriod{
			Kind:  p.Kind,
			Start: p.Start,
			End:   p.End,
		}
	}
	if err := vacation.ValidateCreation(domainPeriods, *matched, vacations, h.workYearCalc); err != nil {
		return nil, err
	}

	// 5. Создание и сохранение всех частей.
	newVacations := make([]*vacation.Vacation, 0, len(domainPeriods))
	ids := make([]string, 0, len(domainPeriods))
	for _, dp := range domainPeriods {
		vacationID := vacation.NewVacationID()
		v, err := vacation.NewVacation(vacationID, userID, dp.Kind, dp.Start, dp.End, time.Now())
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

	return &Result{VacationIDs: ids}, nil
}

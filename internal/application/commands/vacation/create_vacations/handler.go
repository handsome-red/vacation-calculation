package create_vacations

import (
	"context"
	"fmt"
	"time"

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

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	userID, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: некорректный идентификатор пользователя %q", vacation.ErrInvalidInput, cmd.UserID)
	}

	startDate, err := vacation.ParseDate(cmd.Periods[0].StartDate)
	if err != nil {
		return nil, fmt.Errorf("%w: не удалось разобрать дату начала %q", vacation.ErrInvalidInput, cmd.Periods[0].StartDate)
	}

	endDate, err := vacation.ParseDate(cmd.Periods[0].EndDate)
	if err != nil {
		return nil, fmt.Errorf("%w: не удалось разобрать дату окончания %q", vacation.ErrInvalidInput, cmd.Periods[0].EndDate)
	}

	vacationID := vacation.NewVacationID()
	now := time.Now()

	newVacation, err := vacation.NewVacation(vacationID, userID, startDate, endDate, now)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", vacation.ErrInvalidInput, err)
	}

	// 1. Минимальная длительность.
	if days := vacation.DaysInclusive(startDate.Time(), endDate.Time()); days < 7 {
		return nil, fmt.Errorf("%w: запрошено %d дн., минимальная длительность отпуска — 7 дней",
			vacation.ErrTooShort, days)
	}

	u, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	shifts, err := h.shiftRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list shifts: %w", err)
	}

	// 2. Дата начала раньше даты найма.
	hiredAt := u.HiredAt().Time()
	if startDate.Time().Before(hiredAt) {
		return nil, fmt.Errorf("%w: начало отпуска %s раньше даты приёма на работу %s",
			vacation.ErrBeforeHireDate,
			startDate.Time().Format("02.01.2006"),
			hiredAt.Format("02.01.2006"))
	}

	stats, err := h.workYearCalc.FindYearStat(
		hiredAt,
		endDate.Time().AddDate(0, 0, 1),
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

	// 3. Поиск подходящего рабочего года.
	var matched *vacation.YearStat
	for i := range stats {
		year := stats[i]
		if !startDate.Time().Before(year.From) && startDate.Time().Before(year.To) {
			matched = &year
			break
		}
	}
	if matched == nil {
		return nil, fmt.Errorf("%w: дата начала %s не попадает ни в один рабочий год сотрудника",
			vacation.ErrNoSuitableWorkYear,
			startDate.Time().Format("02.01.2006"))
	}

	// 4. Все правила ValidateCreation — уже с конкретными текстами.
	if err := vacation.ValidateCreation(startDate, endDate, *matched, vacations, h.workYearCalc); err != nil {
		return nil, err
	}

	if err := h.vacationRepo.Save(ctx, newVacation); err != nil {
		return nil, fmt.Errorf("failed to save vacation: %w", err)
	}

	return &Result{
		VacationID: vacationID.String(),
	}, nil
}

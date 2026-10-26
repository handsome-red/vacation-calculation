package vacation

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

// Минимальные бизнес-константы, используемые при валидации отпуска.
const (
	// MinVacationDays — минимальная длительность одной части отпуска.
	MinVacationDays = 7

	// LongPartDays — длина непрерывной части отпуска, которая должна быть
	// предусмотрена в течение рабочего года.
	LongPartDays = 14

	// MinGapDays — минимальное количество свободных дней между частями отпуска.
	MinGapDays = 7
)

var (
	ErrBasicVacationRules = errors.New("нарушены правила предоставления отпуска")

	ErrInvalidInput = errors.New("неверный ввод")

	ErrOutsideWorkYear = fmt.Errorf(
		"%w: отпуск должен полностью попадать в границы рабочего года",
		ErrBasicVacationRules)

	ErrTooShort = fmt.Errorf(
		"%w: минимальная длительность отпуска — %d календарных дней",
		ErrBasicVacationRules, MinVacationDays)

	ErrOverlapsExisting = fmt.Errorf(
		"%w: отпуск пересекается с уже оформленным отпуском",
		ErrBasicVacationRules)

	ErrGapTooSmall = fmt.Errorf(
		"%w: между частями отпуска должно быть не менее %d свободных дней",
		ErrBasicVacationRules, MinGapDays)

	ErrExceedsAnnual = fmt.Errorf(
		"%w: длительность отпуска превышает оставшуюся годовую норму дней",
		ErrBasicVacationRules)

	ErrNoLongPart = fmt.Errorf(
		"%w: в течение рабочего года должна быть предусмотрена непрерывная часть отпуска не менее %d дней",
		ErrBasicVacationRules, LongPartDays)

	ErrBeforeHireDate = fmt.Errorf(
		"%w: дата начала отпуска раньше даты приёма на работу",
		ErrBasicVacationRules)

	ErrNoSuitableWorkYear = fmt.Errorf(
		"%w: не удалось определить рабочий год для указанной даты начала отпуска",
		ErrBasicVacationRules)
)

// vacationPeriod — вспомогательная структура для работы с интервалами отпусков.
type vacationPeriod struct {
	from, to time.Time
}

type CreationPeriod struct {
	Kind  Kind
	Start date.Date
	End   date.Date
}

// ValidateCreation проверяет запрошенный период отпуска [start, end] относительно
// рабочего года сотрудника и уже оформленных отпусков.
//
// year.To — исключающая граница (первый день следующего рабочего года).
//
// Возвращает одну из конкретных ошибок (см. блок var выше), обёрнутую через
// fmt.Errorf с человекочитаемым пояснением. Любая из них удовлетворяет
// errors.Is(err, ErrBasicVacationRules).
func ValidateCreation(
	periods []CreationPeriod,
	year YearStat,
	vacations []*Vacation,
	calc WorkYearCalculator,
) error {

	if len(periods) == 0 {
		return fmt.Errorf("%w: не передан ни один период", ErrInvalidInput)
	}

	lastDayOfYear := year.To.AddDate(0, 0, -1)

	newPeriods := make([]vacationPeriod, 0, len(periods))
	totalDays := 0

	for i, p := range periods {
		from, to := p.Start.Time(), p.End.Time()
		if from.Before(year.From) || !to.Before(year.To) {
			return fmt.Errorf(
				"%w: период #%d %s–%s выходит за границы рабочего года %s–%s",
				ErrOutsideWorkYear, i+1,
				from.Format("02.01.2006"), to.Format("02.01.2006"),
				year.From.Format("02.01.2006"), lastDayOfYear.Format("02.01.2006"),
			)
		}

		days := DaysInclusive(from, to)
		if days < MinVacationDays {
			return fmt.Errorf(
				"%w: период #%d: запрошено %d дн., минимум — %d дн.",
				ErrTooShort, i+1, days, MinVacationDays,
			)
		}
		totalDays += days
		newPeriods = append(newPeriods, vacationPeriod{from: from, to: to})
	}

	existingPeriods := make([]vacationPeriod, 0, len(vacations))
	all := make([]vacationPeriod, 0, len(newPeriods)+len(vacations))
	all = append(all, newPeriods...)
	for _, v := range vacations {
		if v == nil || v.Status() == StatusCancelled {
			continue
		}
		vp := vacationPeriod{from: v.StartDate().Time(), to: v.EndDate().Time()}
		existingPeriods = append(existingPeriods, vp)
		all = append(all, vp)
	}

	mergedAll, err := mergePeriods(all)
	if err != nil {
		return err
	}

	mergedExisting, err := mergePeriods(existingPeriods)
	if err != nil {
		return err
	}

	hasLongPart := false
	for _, m := range mergedAll {
		if DaysInclusive(m.from, m.to) >= LongPartDays {
			hasLongPart = true
			break
		}
	}

	annual := year.Base + year.Seniority + year.Irregular
	reserved := 0
	for _, m := range mergedExisting {
		cs := MaxTime(m.from, year.From)
		ce := MinTime(m.to, lastDayOfYear)
		if ce.Before(cs) {
			continue
		}
		reserved += DaysInclusive(cs, ce)
	}

	// ── 6. Разрывы MinGapDays между склеенными интервалами ───────────────
	for i := 1; i < len(mergedAll); i++ {
		gap := int(mergedAll[i].from.Sub(mergedAll[i-1].to).Hours()/24) - 1
		if gap < MinGapDays {
			return fmt.Errorf(
				"%w: между отпусками %s–%s и %s–%s всего %d свободных дн. (нужно %d)",
				ErrGapTooSmall,
				mergedAll[i-1].from.Format("02.01.2006"), mergedAll[i-1].to.Format("02.01.2006"),
				mergedAll[i].from.Format("02.01.2006"), mergedAll[i].to.Format("02.01.2006"),
				gap, MinGapDays,
			)
		}
	}

	// ── 7. Годовая норма ─────────────────────────────────────────────────
	if totalDays > annual-reserved {
		return fmt.Errorf(
			"%w: запрошено %d дн. суммарно, доступно %d из %d (уже зарезервировано %d)",
			ErrExceedsAnnual, totalDays, annual-reserved, annual, reserved,
		)
	}

	// ── 8. Проверка через калькулятор рабочего года ─────────────────────
	earliestStart := newPeriods[0].from
	for _, p := range newPeriods[1:] {
		if p.from.Before(earliestStart) {
			earliestStart = p.from
		}
	}
	usedAtStart := usedAtBefore(mergedExisting, earliestStart)
	info, err := calc.CurrentYear([]YearStat{year}, earliestStart, usedAtStart)
	if err != nil {
		return fmt.Errorf("не удалось рассчитать остаток отпуска: %w", err)
	}
	if float64(totalDays) > info.Remaining {
		return fmt.Errorf(
			"%w: запрошено %d дн. суммарно, остаток по расчёту — %.1f дн.",
			ErrExceedsAnnual, totalDays, info.Remaining,
		)
	}

	// ── 9. Обязательная непрерывная часть ────────────────────────────────
	if !hasLongPart {
		if annual-reserved-totalDays < LongPartDays {
			return fmt.Errorf(
				"%w: после этой заявки по годовой норме останется %d дн. — недостаточно для непрерывной части %d дн.",
				ErrNoLongPart, annual-reserved-totalDays, LongPartDays,
			)
		}
		latestEnd := newPeriods[0].to
		for _, p := range newPeriods[1:] {
			if p.to.After(latestEnd) {
				latestEnd = p.to
			}
		}
		if !hasRoomForLongPart(mergedAll, latestEnd, year.To) {
			return fmt.Errorf(
				"%w: в оставшемся рабочем году нет свободного окна длиной %d дн. с учётом разрывов",
				ErrNoLongPart, LongPartDays,
			)
		}
	}

	return nil
}

// hasRoomForLongPart проверяет, можно ли после даты after разместить в рабочем
// году ещё одну непрерывную часть отпуска длиной не менее LongPartDays,
// соблюдая разрыв MinGapDays с уже запланированными периодами.
//
// periods должен включать новый отпуск и все действующие отпуска сотрудника.
// after — последний день нового отпуска (включительно).
// yearEnd — исключающая граница рабочего года (как year.To).
func hasRoomForLongPart(merged []vacationPeriod, after, yearEnd time.Time) bool {
	cursor := after
	for _, m := range merged {
		// Пропускаем склейки, которые закончились до курсора.
		if m.to.Before(cursor) {
			continue
		}
		// Если курсор внутри/до начала склейки — считаем свободный кусок.
		if m.from.After(cursor) {
			freeEnd := m.from.AddDate(0, 0, -1)
			if DaysInclusive(cursor, freeEnd) >= LongPartDays {
				return true
			}
		}
		// Сдвигаем курсор за склейку.
		if m.to.AddDate(0, 0, 1).After(cursor) {
			cursor = m.to.AddDate(0, 0, 1)
		}
	}
	// Хвост после последней склейки.
	if cursor.Before(yearEnd) {
		freeEnd := yearEnd.AddDate(0, 0, -1)
		if DaysInclusive(cursor, freeEnd) >= LongPartDays {
			return true
		}
	}
	return false
}

// Helpers
func mergePeriods(periods []vacationPeriod) ([]vacationPeriod, error) {
	if len(periods) == 0 {
		return nil, nil
	}

	sorted := make([]vacationPeriod, len(periods))
	copy(sorted, periods)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].from.Equal(sorted[j].from) {
			return sorted[i].to.Before(sorted[j].to)
		}
		return sorted[i].from.Before(sorted[j].from)
	})

	merged := make([]vacationPeriod, 0, len(sorted))

	cur := vacationPeriod{from: sorted[0].from, to: sorted[0].to}
	for _, p := range sorted[1:] {
		switch {
		// Пересечение - ошибка
		case p.from.Before(cur.to) || p.from.Equal(cur.to):
			return nil, fmt.Errorf(
				"%w: период %s–%s пересекается с %s–%s",
				ErrOverlapsExisting,
				p.from.Format("02.01.2006"), p.to.Format("02.01.2006"),
				cur.from.Format("02.01.2006"), cur.to.Format("02.01.2006"),
			)
		// Встык — склеиваем
		case p.from.Equal(cur.to.AddDate(0, 0, 1)):
			cur.to = p.to
		// Разрыв — новый интервал
		default:
			merged = append(merged, cur)
			cur = vacationPeriod{from: p.from, to: p.to}
		}
	}
	merged = append(merged, cur)

	return merged, nil
}

// usedAtBefore — сколько дней из склейки занято строго до момента t
// (в пределах того же интервала — обрезаем по t-1).
func usedAtBefore(merged []vacationPeriod, t time.Time) int {
	used := 0
	for _, m := range merged {
		if !m.from.Before(t) {
			continue
		}
		end := m.to
		if !end.Before(t) {
			end = t.AddDate(0, 0, -1)
		}
		if end.Before(m.from) {
			continue
		}
		used += DaysInclusive(m.from, end)
	}
	return used
}

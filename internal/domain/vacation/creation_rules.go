package vacation

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

// Минимальные бизнес-константы, используемые при валидации отпуска.
const (
	// MinVacationDays — минимальная длительность одной части отпуска.
	MinVacationDays = 7

	// LongPartDays — длина непрерывной части отпуска, которая должна быть
	// предусмотрена в течение рабочего года.
	LongPartDays = 14

	// MinGapDays — минимальное количество свободных дней между частями отпуска.
	MinGapDays = 14

	// MinBalanceAfter — минимальный остаток дней на балансе после оформления отпуска.
	MinBalanceAfter = 7
)

// Ошибки валидации отпуска.
//
// ErrBasicVacationRules — «зонтик»: errors.Is(err, ErrBasicVacationRules)
// сработает для любой из перечисленных ниже ошибок, что позволяет сохранить
// обратную совместимость с вызывающим кодом.
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

	ErrNotEnoughBalance = fmt.Errorf(
		"%w: после оформления отпуска на балансе должно остаться не менее %d дней",
		ErrBasicVacationRules, MinBalanceAfter)

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

// ValidateCreation проверяет запрошенный период отпуска [start, end] относительно
// рабочего года сотрудника и уже оформленных отпусков.
//
// year.To — исключающая граница (первый день следующего рабочего года).
//
// Возвращает одну из конкретных ошибок (см. блок var выше), обёрнутую через
// fmt.Errorf с человекочитаемым пояснением. Любая из них удовлетворяет
// errors.Is(err, ErrBasicVacationRules).
func ValidateCreation(
	start, end Date,
	year YearStat,
	vacations []*Vacation,
	calc WorkYearCalculator,
) error {
	from, to := start.Time(), end.Time()

	// 1. Границы рабочего года. year.To — исключающая граница.
	lastDayOfYear := year.To.AddDate(0, 0, -1)
	if from.Before(year.From) || !to.Before(year.To) {
		return fmt.Errorf(
			"%w: запрошенный период %s–%s выходит за границы рабочего года %s–%s",
			ErrOutsideWorkYear,
			from.Format("02.01.2006"),
			to.Format("02.01.2006"),
			year.From.Format("02.01.2006"),
			lastDayOfYear.Format("02.01.2006"),
		)
	}

	// 2. Минимальная длительность.
	days := DaysInclusive(from, to)
	if days < MinVacationDays {
		return fmt.Errorf(
			"%w: запрошено %d дн., минимум — %d дн.",
			ErrTooShort, days, MinVacationDays,
		)
	}

	annual := year.Base + year.Seniority + year.Irregular
	reserved := 0
	usedAtStart := 0
	hasLongPart := days >= LongPartDays

	// Периоды, среди которых ищем свободные окна: сначала сам новый отпуск,
	// затем все действующие отпуска сотрудника.
	periods := []vacationPeriod{{from: from, to: to}}

	for _, v := range vacations {
		if v == nil || v.Status() == StatusRejected || v.Status() == StatusCancelled {
			continue
		}

		otherStart := v.StartDate().Time()
		otherEnd := v.EndDate().Time()
		periods = append(periods, vacationPeriod{from: otherStart, to: otherEnd})

		// 3. Пересечение с существующим отпуском.
		if !otherEnd.Before(from) && !otherStart.After(to) {
			return fmt.Errorf(
				"%w: пересечение с отпуском %s–%s",
				ErrOverlapsExisting,
				otherStart.Format("02.01.2006"),
				otherEnd.Format("02.01.2006"),
			)
		}

		// 4. Недостаточный разрыв с соседним отпуском.
		// Свободные дни между частями: otherEnd+1 .. from-1 (или to+1 .. otherStart-1).
		if otherEnd.Before(from) {
			gap := int(from.Sub(otherEnd).Hours()/24) - 1
			if gap < MinGapDays {
				return fmt.Errorf(
					"%w: до предыдущего отпуска %s–%s всего %d свободных дн. (нужно %d)",
					ErrGapTooSmall,
					otherStart.Format("02.01.2006"),
					otherEnd.Format("02.01.2006"),
					gap, MinGapDays,
				)
			}
		} else if otherStart.After(to) {
			gap := int(otherStart.Sub(to).Hours()/24) - 1
			if gap < MinGapDays {
				return fmt.Errorf(
					"%w: после запрошенного отпуска до следующего %s–%s всего %d свободных дн. (нужно %d)",
					ErrGapTooSmall,
					otherStart.Format("02.01.2006"),
					otherEnd.Format("02.01.2006"),
					gap, MinGapDays,
				)
			}
		}

		// Накопление зарезервированных дней в рамках рабочего года.
		clippedStart := MaxTime(otherStart, year.From)
		clippedEnd := MinTime(otherEnd, lastDayOfYear)
		if clippedEnd.Before(clippedStart) {
			continue
		}

		reservedDays := DaysInclusive(clippedStart, clippedEnd)
		reserved += reservedDays
		if clippedStart.Before(from) {
			usedAtStart += DaysInclusive(
				clippedStart,
				MinTime(clippedEnd, from.AddDate(0, 0, -1)),
			)
		}
		if reservedDays >= LongPartDays {
			hasLongPart = true
		}
	}

	// 5. Превышение годовой нормы по зарезервированным дням.
	if days > annual-reserved {
		return fmt.Errorf(
			"%w: запрошено %d дн., доступно %d из %d (уже зарезервировано %d)",
			ErrExceedsAnnual,
			days, annual-reserved, annual, reserved,
		)
	}

	// 6. Проверка через калькулятор рабочего года.
	info, err := calc.CurrentYear([]YearStat{year}, from, usedAtStart)
	if err != nil {
		return fmt.Errorf("не удалось рассчитать остаток отпуска: %w", err)
	}
	if float64(days) > info.Remaining {
		return fmt.Errorf(
			"%w: запрошено %d дн., остаток по расчёту — %.1f дн.",
			ErrExceedsAnnual, days, info.Remaining,
		)
	}

	// 6.1. После оформления на балансе должно остаться не менее MinBalanceAfter дней.
	if remainingAfter := info.Remaining - float64(days); remainingAfter < MinBalanceAfter {
		return fmt.Errorf(
			"%w: остаток после оформления — %.1f дн., минимум — %d дн.",
			ErrNotEnoughBalance, remainingAfter, MinBalanceAfter,
		)
	}

	// 7. Обязательная непрерывная часть длиной не менее LongPartDays.
	// Для коротких отпусков (< LongPartDays) проверяем, что её ещё можно
	// разместить в оставшемся рабочем году с учётом свободных окон и разрывов.
	if !hasLongPart {
		if annual-reserved-days < LongPartDays {
			return fmt.Errorf(
				"%w: после этого отпуска по годовой норме останется %d дн. — недостаточно для непрерывной части %d дн.",
				ErrNoLongPart, annual-reserved-days, LongPartDays,
			)
		}
		if !hasRoomForLongPart(periods, to, year.To) {
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
func hasRoomForLongPart(periods []vacationPeriod, after, yearEnd time.Time) bool {
	if len(periods) == 0 {
		return false
	}

	// Работаем по копии, чтобы не мутировать входной слайс.
	sorted := make([]vacationPeriod, len(periods))
	copy(sorted, periods)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].from.Before(sorted[j].from)
	})

	lastDay := yearEnd.AddDate(0, 0, -1)

	// Кандидат — первый день после нового отпуска + обязательный разрыв.
	candidate := after.AddDate(0, 0, MinGapDays+1)

	for _, period := range sorted {
		// Пропускаем периоды, которые закончились до after.
		if !period.to.After(after) {
			continue
		}

		// Если между candidate и началом периода умещается LongPartDays
		// с учётом разрыва MinGapDays до самого периода — окно найдено.
		windowEnd := period.from.AddDate(0, 0, -MinGapDays-1)
		if DaysInclusive(candidate, windowEnd) >= LongPartDays {
			return true
		}

		// Иначе сдвигаем кандидата за конец периода + разрыв.
		if next := period.to.AddDate(0, 0, MinGapDays+1); next.After(candidate) {
			candidate = next
		}
	}

	// Проверяем окно до конца рабочего года.
	return DaysInclusive(candidate, lastDay) >= LongPartDays
}

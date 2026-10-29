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

// ValidateCreation проверяет календарные правила для запрошенных периодов
// относительно уже оформленных отпусков. Достаточность квоты проверяет Allocate.
//
// year — рабочий год, в который попадает начало самого раннего периода;
// year.To — исключающая граница (первый день следующего рабочего года).
//
// violations — нарушения правил предоставления отпуска (длительность, разрывы,
// непрерывная часть). Каждое — обёрнутая ошибка из блока var выше с пояснением,
// удовлетворяет errors.Is(err, ErrBasicVacationRules). Администратор может их обойти.
//
// err — ошибка, которую обойти нельзя: пустой ввод или пересечение с другим отпуском.
func ValidateCreation(
	periods []CreationPeriod,
	year YearStat,
	vacations []*Vacation,
) (violations []error, err error) {

	if len(periods) == 0 {
		return nil, fmt.Errorf("%w: не передан ни один период", ErrInvalidInput)
	}

	newPeriods := make([]vacationPeriod, 0, len(periods))

	for i, p := range periods {
		from, to := p.Start.Time(), p.End.Time()

		days := DaysInclusive(from, to)
		if days < MinVacationDays {
			violations = append(violations, fmt.Errorf(
				"%w: период #%d: запрошено %d дн., минимум — %d дн.",
				ErrTooShort, i+1, days, MinVacationDays,
			))
		}
		newPeriods = append(newPeriods, vacationPeriod{from: from, to: to})
	}

	all := make([]vacationPeriod, 0, len(newPeriods)+len(vacations))
	all = append(all, newPeriods...)
	for _, v := range vacations {
		if v == nil || !v.Status().IsActive() {
			continue
		}
		all = append(all, vacationPeriod{from: v.StartDate().Time(), to: v.EndDate().Time()})
	}

	mergedAll, err := mergePeriods(all)
	if err != nil {
		return nil, err
	}

	hasLongPart := false
	for _, m := range mergedAll {
		if DaysInclusive(m.from, m.to) >= LongPartDays {
			hasLongPart = true
			break
		}
	}

	// ── 6. Разрывы MinGapDays между склеенными интервалами ───────────────
	for i := 1; i < len(mergedAll); i++ {
		gap := int(mergedAll[i].from.Sub(mergedAll[i-1].to).Hours()/24) - 1
		if gap < MinGapDays {
			violations = append(violations, fmt.Errorf(
				"%w: между отпусками %s–%s и %s–%s всего %d свободных дн. (нужно %d)",
				ErrGapTooSmall,
				mergedAll[i-1].from.Format("02.01.2006"), mergedAll[i-1].to.Format("02.01.2006"),
				mergedAll[i].from.Format("02.01.2006"), mergedAll[i].to.Format("02.01.2006"),
				gap, MinGapDays,
			))
		}
	}

	// ── 7. Обязательная непрерывная часть ────────────────────────────────
	if !hasLongPart {
		latestEnd := newPeriods[0].to
		for _, p := range newPeriods[1:] {
			if p.to.After(latestEnd) {
				latestEnd = p.to
			}
		}
		if !hasRoomForLongPart(mergedAll, latestEnd, year.To) {
			violations = append(violations, fmt.Errorf(
				"%w: в оставшемся рабочем году нет свободного окна длиной %d дн. с учётом разрывов",
				ErrNoLongPart, LongPartDays,
			))
		}
	}

	return violations, nil
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

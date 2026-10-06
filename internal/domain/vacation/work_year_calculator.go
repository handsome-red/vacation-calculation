package vacation

import (
	"fmt"
	"sort"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type WorkYearCalculator struct{}

func NewWorkYearCalculator() WorkYearCalculator { return WorkYearCalculator{} }

type YearStat struct {
	From      time.Time // начало рабочего года (включительно)
	To        time.Time // конец рабочего года (исключительно)
	Base      int
	Seniority int
	Irregular int
}

func (c WorkYearCalculator) FindYearStat(
	hiredAt time.Time,
	now time.Time,
	irregular bool,
	shifts []Shift,
	seniorityAt func(at time.Time) int,
) []YearStat {
	if !hiredAt.Before(now) {
		return nil
	}
	if err := validateShifts(shifts); err != nil {
		panic(err) // программная ошибка, данные должны быть валидны
	}

	result := make([]YearStat, 0)
	cursor := hiredAt

	for cursor.Before(now) {
		yearStart := cursor
		yearEnd := yearStart.AddDate(1, 0, 0)
		yearEnd = yearEnd.AddDate(0, 0, shiftDaysForYear(yearStart, yearEnd, shifts))

		asOf := yearEnd
		if asOf.After(now) {
			asOf = now
		}

		result = append(result, YearStat{
			From:      yearStart,
			To:        yearEnd,
			Base:      BaseAnnualDays,
			Seniority: seniorityAt(asOf),
			Irregular: irregularDays(irregular),
		})

		cursor = yearEnd
	}

	return result
}

func (c WorkYearCalculator) SeniorityAtFromHired(hiredAt time.Time) func(at time.Time) int {
	return func(at time.Time) int {
		return seniorityForExperience(fullYears(hiredAt, at))
	}
}

func seniorityForExperience(years int) int {
	switch {
	case years < 1:
		return 0
	case years < 5:
		return 1
	case years < 10:
		return 5
	case years < 15:
		return 7
	default:
		return 10
	}
}

func validateShifts(shifts []Shift) error {
	sorted := make([]Shift, len(shifts))
	copy(sorted, shifts)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].From.Before(sorted[j].From)
	})

	for i, s := range sorted {
		if s.To.Before(s.From) {
			return fmt.Errorf("shift[%d]: To (%s) раньше From (%s)", i, s.To, s.From)
		}
		if _, ok := shiftMap[s.Kind]; !ok {
			return fmt.Errorf("shift[%d]: неизвестный Kind: %s", i, s.Kind)
		}
		if i > 0 && !s.From.After(sorted[i-1].To) {
			return fmt.Errorf(
				"shift[%d] пересекается с shift[%d]: %s..%s и %s..%s",
				i-1, i,
				sorted[i-1].From, sorted[i-1].To,
				s.From, s.To,
			)
		}
	}
	return nil
}

func irregularDays(on bool) int {
	if on {
		return IrregularDays
	}
	return 0
}

type Interval struct {
	from, to time.Time
}

func AnnualEntitlement(hired user.HiredDate, irregular bool, now time.Time) int {
	total := baseAllowance() + seniorityAllowance(hired, now)
	if irregular {
		total += IrregularDays
	}
	return total
}

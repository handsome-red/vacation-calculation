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
	Shifts    []Shift
}

type CurrentYearInfo struct {
	Year      YearStat
	Earned    float64
	Used      int
	Remaining float64
	Today     time.Time
}

func (c WorkYearCalculator) FindYearStat(
	hiredAt time.Time,
	now time.Time,
	irregular bool,
	shifts []Shift,
	seniorityAt func(at time.Time) int,
) ([]YearStat, error) {
	if !hiredAt.Before(now) {
		return nil, ErrCalcInvalidTime
	}
	if err := validateShifts(shifts); err != nil {
		return nil, fmt.Errorf("validate shifts: %w", err)
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
			Shifts:    shiftsOverlapping(shifts, yearStart, yearEnd),
		})

		cursor = yearEnd
	}

	return result, nil
}

// Получаем статистику по последнему рабочему году - текущему
func (c WorkYearCalculator) GetLastPeriod(stats []YearStat) YearStat {
	return stats[len(stats)-1]
}

// Использовано отпускных
// func (c WorkYearCalculator) UsedVacationDays(from, to time.Time, vacation []Vacation) int {
// 	used := 0
// 	for _, v := range vacation {
// 		used += DaysInclusive(v.StartDate().Time(), v.EndDate().Time())
// 	}

// }

// // Смотрим сколько осталось отпускных
// func (c WorkYearCalculator) FindRemaining(now time.Time, stat YearStat, vacation []Vacation) int {

// }

// func DaysInclusive(start, end time.Time) int {
// 	start = start.UTC()
// 	end = end.UTC()
// 	days := int(end.Sub(start).Hours()/24) + 1
// 	if days < 0 {
// 		return 0
// 	}
// 	return days
// }

func CurrentYear(stats []YearStat, now time.Time) CurrentYearInfo {
	info := CurrentYearInfo{Today: time.Time{}}
	// y := currentYear(stats, now)
	// info.Year = y
	return info
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

func shiftsOverlapping(shifts []Shift, from, to time.Time) []Shift {
	out := make([]Shift, 0)
	for _, s := range shifts {
		if s.From.Before(to) && s.To.After(from) {
			out = append(out, s)
		}
	}
	return out
}

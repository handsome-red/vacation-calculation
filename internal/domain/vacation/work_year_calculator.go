package vacation

import (
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type WorkYearCalculator struct{}

func NewWorkYearCalculator() *WorkYearCalculator { return &WorkYearCalculator{} }

type YearStat struct {
	Year      int
	Base      int
	Seniority int
	Irregular int
}

func FindYearStat(start, end time.Time, experience int, irregular bool) []YearStat {
	result := make([]YearStat, 0)
	for cursor := start; cursor.Before(end); cursor = cursor.AddDate(1, 0, 0) {
		result = append(result, YearStat{
			Year:      cursor.Year(),
			Base:      BaseAnnualDays,
			Seniority: seniorityForExperience(experience),
			Irregular: irregularDays(irregular),
		})
	}
	return result
}

type Interval struct {
	from, to time.Time
}

func (c WorkYearCalculator) Calculate(
	hiredAt time.Time,
	now time.Time,
	interval []Interval,
) []Interval {
	if !hiredAt.Before(now) {
		return nil
	}

	result := make([]Interval, 0)
	// for start := hiredAt; start.Before(now); {
	// 	end := start.AddDate(1, 0, 0)

	// 	if
	// }

	return result
}

func AnnualEntitlement(hired user.HiredDate, irregular bool, now time.Time) int {
	total := baseAllowance() + seniorityAllowance(hired, now)
	if irregular {
		total += IrregularDays
	}
	return total
}

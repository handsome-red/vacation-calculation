package vacation

import "time"

type WorkYearCalculator struct{}

func NewWorkYearCalculator() *WorkYearCalculator { return &WorkYearCalculator{} }

type YearStat struct {
	Year      int
	Base      int
	Seniotiry int
	Irregular int
}

func FindYearStat(start, end time.Time, experience int) []YearStat {
	result := make([]YearStat, 0)
	for start.Before(end) {
		result = append(result, YearStat{
			Year: start.Year(),
		})

		start = start.AddDate(1, 0, 0)
	}
	return result
}

type Interval struct{
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
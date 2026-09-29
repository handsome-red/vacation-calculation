package vacation

import "time"

type WorkYearCalculator struct{}

type Shift struct {
	from, to time.Time
	reason   string
}

func NewWorkYearCalculator() *WorkYearCalculator { return &WorkYearCalculator{} }

func (s Shift) Days() int {
	return s.to.Day() - s.from.Day()
}

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

func (c WorkYearCalculator) Calculate(
	start time.Time,
	end time.Time,
	shifts []Shift,
) (from, to time.Time) {
	from = start
	for !end.Before(from.AddDate(1, 0, 0)) {
		from = from.AddDate(1, 0, 0)
	}

	to = from.AddDate(1, 0, 0)
	for _, shift := range shifts {
		if !shift.from.Before(from) && !shift.to.After(to) {
			diff := to.Sub(from)
			to = to.Add(diff)
		}
	}

	return from, to
}

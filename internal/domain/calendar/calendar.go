package calendar

import (
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

type Calendar struct {
	Year  int
	Month []Month
}

type Month struct {
	Number time.Month
	Days   []Day
}

type Day struct {
	Date date.Date
	Kind DayType
	Name string
}

func BuildYear(year int, exceptions map[date.Date]Exception) (Calendar, error) {
	months := make([]Month, 0, 12)

	for m := 1; m <= 12; m++ {
		month := time.Month(m)
		daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
		days := make([]Day, 0, daysInMonth)

		for d := 1; d <= daysInMonth; d++ {
			d := date.New(year, month, d)

			e := exceptions[d]
			days = append(days, Day{
				Date: d,
				Kind: e.DayType,
				Name: e.Name,
			})
		}

		months = append(months, Month{
			Number: month,
			Days:   days,
		})
	}

	return Calendar{
		Year:  year,
		Month: months,
	}, nil
}

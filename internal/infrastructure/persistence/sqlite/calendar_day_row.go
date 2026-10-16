package sqlite

import (
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

type calendarDayRow struct {
	Date    string `db:"date"`
	DayType string `db:"day_type"`
	Name    string `db:"name"`
}

func (r calendarDayRow) toDomain() (date.Date, error) {
	date, err := date.ParseDate(r.Date)
	if err != nil {
		return date.Date{}, fmt.Errorf("parse date: %w", err)
	}
	return date.NewCalendarDay(date, date.DateType(r.DayType), r.Name)
}

package sqlite

import (
	"github.com/handsome-red/vacation-calculation/internal/domain/calendar"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

type calendarDayRow struct {
	Date    string `db:"date"`
	DayType string `db:"day_type"`
	Name    string `db:"name"`
}

func (r calendarDayRow) toDomain() (calendar.Exception, error) {
	d, err := date.ParseDate(r.Date)
	if err != nil {
		return calendar.Exception{}, err
	}
	dt, err := calendar.ParseDayType(r.DayType)
	if err != nil {
		return calendar.Exception{}, err
	}
	return calendar.NewException(d, dt, r.Name)
}

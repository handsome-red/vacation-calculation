package sqlite

import (
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type holidayRow struct {
	Name string `db:"name"`
	Date string `db:"date"`
}

func (r holidayRow) toDomain() (vacation.Holiday, error) {
	date, err := vacation.ParseDate(r.Date)
	if err != nil {
		return vacation.Holiday{}, fmt.Errorf("invalid holiday date: %q, %w", r.Date, err)
	}

	h, err := vacation.NewHoliday(r.Name, date)
	if err != nil {
		return vacation.Holiday{}, fmt.Errorf("invalid holiday: %q, %w", r.Date, err)
	}

	return h, nil
}
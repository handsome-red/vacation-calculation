package sqlite

import (
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type holidayRow struct {
	Name string        `db:"name"`
	Date vacation.Date `db:"date"`
}

func (r holidayRow) toDomain() (vacation.Holiday, error) {
	h, err := vacation.NewHoliday(r.Name, r.Date)
	if err != nil {
		return vacation.Holiday{}, fmt.Errorf("holiday %q: %w", r.Date.ISO(), err)
	}
	return h, nil
}

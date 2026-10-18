package calendar

import (
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

type Exception struct {
	date    date.Date
	dayType DayType
	name    string
}

func NewException(d date.Date, t DayType, name string) (Exception, error) {
	if d.IsZero() {
		return Exception{}, fmt.Errorf("calendar exception: empty date")
	}
	return Exception{date: d, dayType: t, name: name}, nil
}

func ParseDayType(s string) (DayType, error) {
    t := DayType(s)
    if !t.Valid() {
        return "", fmt.Errorf("calendar: unknown day_type %q", s)
    }
    return t, nil
}

func (e Exception) Date() date.Date   { return e.date }
func (e Exception) DayType() DayType  { return e.dayType }
func (e Exception) Name() string      { return e.name }
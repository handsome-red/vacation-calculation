package vacation

import "time"

const dateLayout = "02.01.2006"

type Date struct {
	year  int
	month time.Month
	day   int
}

func (d Date) IsZero() bool {
	panic("unimplemented")
}

func (d Date) Before(start Date) bool {
	return d.Time().Before(start.Time())
}

func NewDate(y int, m time.Month, d int) (Date, error) {
	t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	if t.Year() != y || t.Month() != m || t.Day() != d {
		return Date{}, ErrInvalidDate
	}
	return Date{
		year:  y,
		month: m,
		day:   d,
	}, nil
}

func (d Date) Year() int         { return d.year }
func (d Date) Month() time.Month { return d.month }
func (d Date) Day() int          { return d.day }

func (d Date) Time() time.Time {
	return time.Date(d.year, d.month, d.day, 0, 0, 0, 0, time.UTC)
}

func ParseDate(s string) (Date, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return Date{}, ErrInvalidDate
	}
	return Date{
		year:  t.Year(),
		month: t.Month(),
		day:   t.Day(),
	}, nil
}
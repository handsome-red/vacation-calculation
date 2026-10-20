package date

import (
	"database/sql/driver"
	"fmt"
	"log"
	"time"
)

const dateLayout = "2006-01-02"

var DateFormat = "02.01.2006"

type Date struct {
	year  int
	month time.Month
	day   int
}

func New(
	year int,
	month time.Month,
	day int,
) Date {
	return Date{
		year:  year,
		month: month,
		day:   day,
	}
}

func (d Date) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.ISO(), nil
}

func (d *Date) Scan(src any) error {
	if src == nil {
		*d = Date{}
		return nil
	}

	var s string
	switch v := src.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	case time.Time:
		*d = Date{year: v.Year(), month: v.Month(), day: v.Day()}
		return nil
	default:
		return fmt.Errorf("date.Date: cannot scan type %T", src)
	}

	parsed, err := ParseDate(s)
	if err != nil {
		return fmt.Errorf("date.Date: scan %q: %w", s, err)
	}
	*d = parsed
	return nil
}

func (d Date) IsZero() bool {
	return d.year == 0 && d.month == 0 && d.day == 0
}

func (d Date) Before(start Date) bool {
	return d.Time().Before(start.Time())
}

func (d Date) After(start Date) bool {
	return d.Time().After(start.Time())
}

func (d Date) String() string {
	return fmt.Sprintf("%d.%d.%d", d.day, d.month, d.year)
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

func (d Date) AddDay() Date {
	result := d.Time().AddDate(0, 0, 1)
	return New(result.Year(), result.Month(), result.Day())
}

func ParseDate(s string) (Date, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		log.Printf("ParseDate: input=%q err=%v", s, err)
		return Date{}, ErrInvalidDate
	}
	return Date{
		year:  t.Year(),
		month: t.Month(),
		day:   t.Day(),
	}, nil
}

func (d Date) ISO() string {
	if d.IsZero() {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", d.year, int(d.month), d.day)
}

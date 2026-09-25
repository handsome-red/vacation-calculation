package user

import (
	"fmt"
	"strings"
	"time"
)

// TODO
type HiredDate struct{ value time.Time }
type WorkYear struct{ from, to time.Time }

func NewHiredDate(s string) (HiredDate, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return HiredDate{}, ErrHiredDateInvalid
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return HiredDate{}, fmt.Errorf("%w: %v", ErrHiredDateInvalid, err)
	}
	t = t.UTC()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	if t.After(today) || t.Year() < 1900 {
		return HiredDate{}, ErrHiredDateInvalid
	}
	return HiredDate{value: t}, nil
}

func (h HiredDate) Time() time.Time {
	return h.value
}

func (h HiredDate) String() string {
	if h.value.IsZero() {
		return ""
	}
	return h.value.UTC().Format("2006-01-02")
}

// WorkYear - возвращает строковое представление рабочего года сотрудника
func (h HiredDate) WorkYear(now time.Time) WorkYear {
	years := now.Year() - h.value.Year()

	from := h.value.AddDate(years, 0, 0)
	if from.After(now) {
		from = from.AddDate(-1, 0, 0)
	}

	to := from.AddDate(1, 0, 0)
	return WorkYear{
		from: from,
		to:   to,
	}
}

// TODO: Реализовать более элегантно
func (w WorkYear) Year(now time.Time) int {
	years := w.to.Year() - w.from.Year()
	if w.from.AddDate(years, 0, 0).After(now) {
		years--
	}
	
	return years
}

func (w WorkYear) String() string {

	f := w.from.UTC().Format("02.01.2006")
	t := w.to.UTC().Format("02.01.2006")

	return fmt.Sprintf("%s - %s", f, t)
}

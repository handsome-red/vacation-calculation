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
	return h.value.UTC().Format("02.01.2006")
}

func (w WorkYear) String() string {

	f := w.from.UTC().Format("02.01.2006")
	t := w.to.UTC().Format("02.01.2006")

	return fmt.Sprintf("%s - %s", f, t)
}

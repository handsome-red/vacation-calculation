package user

import (
	"fmt"
	"strings"
	"time"
)

// TODO
type BirthDate struct{ value time.Time }

func NewBirthDate(s string) (BirthDate, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return BirthDate{}, ErrBirthDateInvalid
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return BirthDate{}, fmt.Errorf("%w: %v", ErrBirthDateInvalid, err)
	}
	t = t.UTC()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	if t.After(today) || t.Year() < 1900 {
		return BirthDate{}, ErrBirthDateInvalid
	}
	return BirthDate{value: t}, nil
}

func (b BirthDate) String() string {
	if b.value.IsZero() {
		return ""
	}
	return b.value.Format("2006-01-02")
}

func (b BirthDate) HumanRead() string {
	if b.value.IsZero() {
		return ""
	}

	years := fullYears(b.value, time.Now())
	date := b.value.UTC().Format("2006-01-02")

	return fmt.Sprintf("%s (полных: %dг./л.)", date, years)
}

func fullYears(from, to time.Time) int {
	years := to.Year() - from.Year()
	if from.AddDate(years, 0, 0).After(to) {
		years--
	}
	return years
}

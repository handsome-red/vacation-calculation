package user

import (
	"time"
)

// TODO
type BirthDate struct {
	value time.Time
}

func NewBirthDate(value time.Time) (BirthDate, error) {

	now := time.Now()

	if value.After(now) {
		return BirthDate{}, ErrBirthDateInvalid
	}

	return BirthDate{value: value}, nil
}

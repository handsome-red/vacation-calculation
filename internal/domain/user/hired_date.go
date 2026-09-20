package user

import (
	"time"
)

// TODO
type HiredDate struct {
	value time.Time
}

func NewHiredDate(value time.Time) (HiredDate, error) {

	now := time.Now()

	if value.After(now) {
		return HiredDate{}, ErrHiredDateInvalid
	}

	return HiredDate{value: value}, nil
}

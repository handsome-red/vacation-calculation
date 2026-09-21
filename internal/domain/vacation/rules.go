package vacation

import (
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

const (
	BaseAnnualDays = 30
	IggerularDays  = 3
)

// Количество дней отпуска за выслугу лет
func baseAllowance() int {
	return BaseAnnualDays
}

// Количество дней отпуска за выслугу лет
// TODO: Считается ли только в Роспотребнадзоре или же в других госорганах тоже?
func seniorityAllowance(hired user.HiredDate, now time.Time) int {

	years := fullYears(hired.Time(), now)

	switch {
	case years < 1:
		return 0
	case years > 0 && years < 5:
		return 1
	case years > 5 && years < 10:
		return 5
	case years > 10 && years < 15:
		return 7
	default:
		return 10
	}
}

func fullYears(from, to time.Time) int {
	if from.IsZero() || to.Before(from) {
		return 0
	}

	years := to.Year() - from.Year()

	if to.YearDay() < from.YearDay() {
		years--
	}

	return years
}

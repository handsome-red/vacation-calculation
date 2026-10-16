package calendar

import (
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

type Exception struct {
	Date    date.Date
	DayType DayType
	Name    string
}

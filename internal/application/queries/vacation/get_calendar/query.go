package get_calendar

import (
	"github.com/handsome-red/vacation-calculation/internal/domain/calendar"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

type Query struct {
	Year int
}

type EventKind string

const (
	EventVacation EventKind = "vacation"
	EventShift    EventKind = "shift"
)

type Event struct {
	Kind     EventKind // "vacation" | "shift"
	SubKind  string    // для сдвига: "unpaid_long" | "parental_leave" | "absenteeism"; для отпуска — статус
	Title    string    // "Отпуск 14 дн." / "Прогул 3 дн."
	From, To date.Date
}

type Result struct {
	Calendar calendar.Calendar
}

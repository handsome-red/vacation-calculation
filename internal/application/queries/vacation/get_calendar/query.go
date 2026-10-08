package get_calendar

import "github.com/handsome-red/vacation-calculation/internal/domain/vacation"

type Query struct {
	UserID string
	Year   int
	Month  int
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
	From, To vacation.Date
}

type Day struct {
	Date        vacation.Date
	IsWeekend   bool
	IsHoliday   bool
	HolidayName string
	Events      []Event
}

type Result struct {
	Year  int
	Month int
	Days  []Day
}

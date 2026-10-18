package get_user_calendar

import "github.com/handsome-red/vacation-calculation/internal/domain/date"

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
	From, To date.Date
}

type Day struct {
	Date             date.Date `json:"date"`
	IsWeekend        bool      `json:"isWeekend"`
	IsHoliday        bool      `json:"isHoliday"`
	HolidayName      string    `json:"holidayName,omitempty"`
	IsWorkSaturday   bool      `json:"isWorkSaturday"`
	WorkSaturdayName string    `json:"workSaturdayName,omitempty"`
}

type Result struct {
	Year  int
	Month int
	Days  []Day
}

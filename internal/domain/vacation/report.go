package vacation

import "time"

type ReportRow struct {
	FullName     string
	Position     string
	VacationDays int
	StartDate    time.Time
	EndDate      time.Time
	Status       string
}

type Report struct {
	GeneratedAt time.Time
	PeriodFrom  time.Time
	PeriodTo    time.Time
	Rows        []ReportRow
}

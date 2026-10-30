package ports

// ScheduleReport — данные графика отпусков на календарный год для выгрузки в Word.
type ScheduleReport struct {
	Year        int
	Departments []string // отделы сотрудников, вошедших в график
	Rows        []ScheduleReportRow
}

// ScheduleReportRow — строка графика: сотрудник и рабочий год, за который предоставляется отпуск.
type ScheduleReportRow struct {
	Position string
	FullName string

	BaseDays       int
	BaseDates      []string // "01.06.2027 — 14.06.2027"
	IrregularDays  int
	IrregularDates []string
	SeniorityDays  int
	SeniorityDates []string

	WorkPeriod string // "27.01.2025-26.01.2026"
}

// ScheduleRenderer формирует документ графика отпусков.
type ScheduleRenderer interface {
	RenderSchedule(r ScheduleReport) ([]byte, error)
}

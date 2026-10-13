package vacation

import "time"

type ReportRow struct {
	Position string
	FullName string

	BaseDays  int
	BaseDates []time.Time

	IrregularDays  int
	IrregularDates []time.Time

	SeniorityDays  int
	SeniorityDates []time.Time

	WorkYearFrom time.Time
	WorkYearTo   time.Time
}

type Report struct {
	Year            int
	DistrictTitle   string
	DepartmentTitle string

	Approver Approver
	Rows     []ReportRow
}

type Approver struct {
	Position string
	FullName string
	Date     time.Time
}

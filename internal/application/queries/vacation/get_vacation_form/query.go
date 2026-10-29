package get_vacation_form

type Query struct {
	UserID string
}

type Quota struct {
	Total   int
	Blocked int
}

type YearStat struct {
	Year      string
	Start     string // начало рабочего года, ISO — ключ для сопоставления с распределением
	Base      Quota
	Seniority Quota
	Irregular Quota
}

type Result struct {
	UserID        string
	Total         int
	Used          int
	Available     int
	HasActiveYear bool
	YearStats     []YearStat
	// Warnings — предупреждения о сроке переноса неиспользованного отпуска.
	Warnings []string
}

package vacation

import "github.com/handsome-red/vacation-calculation/internal/domain/date"

// OpeningDays — дни рабочего года, использованные до начала учёта в системе.
type OpeningDays struct {
	Base      int
	Irregular int
	Seniority int
}

// OpeningUsage — использованные дни по рабочему году. Даты таких отпусков неизвестны,
// поэтому они уменьшают остаток квоты, но не занимают дни календаря.
type OpeningUsage struct {
	WorkYearStart date.Date
	Days          OpeningDays
}

// ApplyOpeningUsage проставляет использованные дни рабочим годам с тем же началом.
func ApplyOpeningUsage(stats []YearStat, usages []OpeningUsage) {
	byYear := make(map[date.Date]OpeningDays, len(usages))
	for _, u := range usages {
		byYear[u.WorkYearStart] = u.Days
	}
	for i := range stats {
		stats[i].Opening = byYear[DateOf(stats[i].From)]
	}
}

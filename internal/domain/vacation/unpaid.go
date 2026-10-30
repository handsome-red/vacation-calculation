package vacation

import (
	"strconv"
	"time"
)

// UnpaidFreeDays — дни отпуска без сохранения заработной платы за рабочий год,
// которые не сдвигают рабочий год (ст. 121 ТК РФ).
const UnpaidFreeDays = 14

// WithUnpaidLeaves добавляет к исключаемым периодам утверждённые административные
// отпуска: как и неоплачиваемый отпуск из сдвигов, они сдвигают рабочий год
// на дни сверх UnpaidFreeDays.
func WithUnpaidLeaves(shifts []Shift, vacations []*Vacation) []Shift {
	out := make([]Shift, 0, len(shifts))
	out = append(out, shifts...)
	for _, v := range vacations {
		if v != nil && v.Kind() == KindUnpaid && v.Status() == StatusApproved {
			out = append(out, NewShift(ShiftKindUnpaid, v.StartDate().Time(), v.EndDate().Time()))
		}
	}
	return out
}

// UnpaidWarning — за рабочий год оформлено больше UnpaidFreeDays дней без сохранения
// заработной платы; лишние дни сдвинули рабочий год.
type UnpaidWarning struct {
	YearFrom time.Time
	YearTo   time.Time
	Days     int
	Shifted  int
}

// UnpaidWarnings — по рабочим годам, где дней без сохранения заработной платы
// (административные отпуска и неоплачиваемые отпуска из сдвигов) больше 14.
func UnpaidWarnings(stats []YearStat, shifts []Shift) []UnpaidWarning {
	out := make([]UnpaidWarning, 0)
	for _, y := range stats {
		lastDay := y.To.AddDate(0, 0, -1)
		days := 0
		for _, s := range shifts {
			if s.Kind != ShiftKindUnpaid {
				continue
			}
			from, to := MaxTime(s.From, y.From), MinTime(s.To, lastDay)
			if !to.Before(from) {
				days += DaysInclusive(from, to)
			}
		}
		if days > UnpaidFreeDays {
			out = append(out, UnpaidWarning{YearFrom: y.From, YearTo: y.To, Days: days, Shifted: days - UnpaidFreeDays})
		}
	}
	return out
}

func (w UnpaidWarning) Message() string {
	return "За рабочий год " + yearRange(w.YearFrom, w.YearTo) + " оформлено " + daysRu(w.Days) +
		" отпуска без сохранения заработной платы (больше " + strconv.Itoa(UnpaidFreeDays) +
		"): рабочий год сдвинут на " + daysRu(w.Shifted) + "."
}

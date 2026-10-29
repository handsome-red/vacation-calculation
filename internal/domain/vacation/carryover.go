package vacation

import (
	"sort"
	"strconv"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

// CarryoverDeadline — последний день, когда можно использовать отпуск за рабочий
// год: не позднее 12 месяцев после его окончания (ст. 124 ТК РФ).
// year.To — первый день следующего рабочего года, поэтому срок — year.To + 12 мес. − 1 день.
func CarryoverDeadline(year YearStat) date.Date {
	return DateOf(year.To.AddDate(1, 0, -1))
}

// CarryoverWarning — неиспользованные дни закончившегося рабочего года.
type CarryoverWarning struct {
	YearFrom time.Time
	YearTo   time.Time // исключительно
	Unused   int
	Deadline date.Date
	Expired  bool // срок использования уже прошёл
}

// CarryoverWarnings — по закончившимся к today рабочим годам, где остались
// неиспользованные (и не запланированные) дни.
func CarryoverWarnings(balances []YearBalance, today date.Date) []CarryoverWarning {
	out := make([]CarryoverWarning, 0)
	for _, b := range balances {
		if today.Time().Before(b.Year.To) || b.Total() == 0 {
			continue
		}
		deadline := CarryoverDeadline(b.Year)
		out = append(out, CarryoverWarning{
			YearFrom: b.Year.From,
			YearTo:   b.Year.To,
			Unused:   b.Total(),
			Deadline: deadline,
			Expired:  today.After(deadline),
		})
	}
	return out
}

// LateUse — отпуск (кусок), использующий дни рабочего года позже срока переноса.
type LateUse struct {
	Start    date.Date
	End      date.Date
	YearFrom time.Time
	YearTo   time.Time
	Deadline date.Date
}

// LateUses — действующие отпуска, которые заканчиваются позже срока использования
// дней своего рабочего года-источника.
func LateUses(vacations []*Vacation, stats []YearStat) []LateUse {
	years := make(map[date.Date]YearStat, len(stats))
	for _, s := range stats {
		years[DateOf(s.From)] = s
	}

	out := make([]LateUse, 0)
	for _, v := range vacations {
		if v == nil || !v.Status().IsActive() {
			continue
		}
		y, ok := years[v.WorkYearStart()]
		if !ok {
			continue
		}
		if deadline := CarryoverDeadline(y); v.EndDate().After(deadline) {
			out = append(out, LateUse{Start: v.StartDate(), End: v.EndDate(), YearFrom: y.From, YearTo: y.To, Deadline: deadline})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Start.Before(out[j].Start) })
	return out
}

// LateAllocations — куски распределения, использующие дни года позже срока переноса.
func LateAllocations(allocations []Allocation, balances []YearBalance) []LateUse {
	years := make(map[date.Date]YearStat, len(balances))
	for _, b := range balances {
		years[DateOf(b.Year.From)] = b.Year
	}

	out := make([]LateUse, 0)
	for _, a := range allocations {
		y, ok := years[a.WorkYearStart]
		if !ok {
			continue
		}
		if deadline := CarryoverDeadline(y); a.End.After(deadline) {
			out = append(out, LateUse{Start: a.Start, End: a.End, YearFrom: y.From, YearTo: y.To, Deadline: deadline})
		}
	}
	return out
}

// CarryoverMessages — все предупреждения о сроке переноса на дату today:
// неиспользованные дни закончившихся лет и отпуска, запланированные позже срока.
func CarryoverMessages(balances []YearBalance, vacations []*Vacation, stats []YearStat, today date.Date) []string {
	out := make([]string, 0)
	for _, w := range CarryoverWarnings(balances, today) {
		out = append(out, w.Message())
	}
	for _, l := range LateUses(vacations, stats) {
		out = append(out, l.Message())
	}
	return out
}

// Тексты предупреждений — общие для карточки, формы отпуска и предпросмотра.

func (w CarryoverWarning) Message() string {
	period := yearRange(w.YearFrom, w.YearTo)
	if w.Expired {
		return "За рабочий год " + period + " не использовано " + daysRu(w.Unused) +
			". Срок переноса истёк " + w.Deadline.Time().Format("02.01.2006") + "."
	}
	return "За рабочий год " + period + " не использовано " + daysRu(w.Unused) +
		". Их нужно использовать не позднее " + w.Deadline.Time().Format("02.01.2006") + "."
}

func (l LateUse) Message() string {
	return "Отпуск " + l.Start.Time().Format("02.01.2006") + " — " + l.End.Time().Format("02.01.2006") +
		" использует дни рабочего года " + yearRange(l.YearFrom, l.YearTo) +
		", которые нужно использовать не позднее " + l.Deadline.Time().Format("02.01.2006") + "."
}

func yearRange(from, to time.Time) string {
	return from.Format("02.01.2006") + "–" + to.AddDate(0, 0, -1).Format("02.01.2006")
}

func daysRu(n int) string {
	word := "дней"
	switch m := n % 100; {
	case m >= 11 && m <= 14:
	case m%10 == 1:
		word = "день"
	case m%10 >= 2 && m%10 <= 4:
		word = "дня"
	}
	return strconv.Itoa(n) + " " + word
}

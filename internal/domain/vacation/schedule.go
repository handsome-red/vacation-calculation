package vacation

import (
	"fmt"
	"sort"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

// SchedulePart — часть ежегодного отпуска в графике: записи, идущие встык
// (дни разных рабочих лет или видов), склеены в одну часть.
type SchedulePart struct {
	Start    date.Date
	End      date.Date
	Days     int  // дни отпуска (без праздников)
	Approved bool // все записи части утверждены
}

// ApprovalDeadline — крайний срок утверждения графика на год: за 2 недели
// до начала календарного года (17 декабря предыдущего).
func ApprovalDeadline(year int) date.Date {
	return date.New(year-1, time.December, 17)
}

// ScheduleParts — части действующего ежегодного отпуска, начинающиеся в календарном году year.
func ScheduleParts(vacations []*Vacation, year int, h Holidays) []SchedulePart {
	active := make([]*Vacation, 0, len(vacations))
	for _, v := range vacations {
		if v != nil && v.Status().IsActive() && v.Kind().IsAnnual() {
			active = append(active, v)
		}
	}
	sort.Slice(active, func(i, j int) bool { return active[i].StartDate().Before(active[j].StartDate()) })

	parts := make([]SchedulePart, 0)
	for _, v := range active {
		approved := v.Status() == StatusApproved
		if n := len(parts); n > 0 && parts[n-1].End.AddDay() == v.StartDate() {
			parts[n-1].End = v.EndDate()
			parts[n-1].Approved = parts[n-1].Approved && approved
			continue
		}
		parts = append(parts, SchedulePart{Start: v.StartDate(), End: v.EndDate(), Approved: approved})
	}

	out := make([]SchedulePart, 0, len(parts))
	for _, p := range parts {
		if p.Start.Year() == year {
			p.Days = ChargeableDays(p.Start, p.End, h)
			out = append(out, p)
		}
	}
	return out
}

// ScheduleIssues — нарушения графика сотрудника на год: нет отпуска, нет части
// не менее 14 дней, части короче 7 дней, перерывы короче 14 дней.
func ScheduleIssues(parts []SchedulePart) []string {
	if len(parts) == 0 {
		return []string{"Отпуск в графике не запланирован"}
	}

	issues := make([]string, 0)
	hasLong := false
	for _, p := range parts {
		if p.Days >= LongPartDays {
			hasLong = true
		}
		if p.Days < MinVacationDays {
			issues = append(issues, fmt.Sprintf("Часть %s — %s: %d дн., минимум %d", fmtD(p.Start), fmtD(p.End), p.Days, MinVacationDays))
		}
	}
	if !hasLong {
		issues = append(issues, fmt.Sprintf("Нет части отпуска не менее %d календарных дней", LongPartDays))
	}
	for i := 1; i < len(parts); i++ {
		gap := int(parts[i].Start.Time().Sub(parts[i-1].End.Time()).Hours()/24) - 1
		if gap < MinGapDays {
			issues = append(issues, fmt.Sprintf("Перерыв между %s и %s — %d дн., нужно не менее %d",
				fmtD(parts[i-1].End), fmtD(parts[i].Start), gap, MinGapDays))
		}
	}
	return issues
}

func fmtD(d date.Date) string { return d.Time().Format("02.01.2006") }

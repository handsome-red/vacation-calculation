package vacation

import (
	"fmt"
	"strings"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

// Discrepancies — расхождения в учёте отпусков сотрудника: рабочие годы, из которых
// списано больше нормы по виду отпуска, и действующие отпуска, привязанные к рабочему
// году, которого нет среди stats. Такие данные остались от времени, когда норму
// можно было превысить; новые отпуска их не создают. Отпуска, привязанные к годам
// позже последнего из stats, не проверяются: эти годы ещё не посчитаны.
func Discrepancies(stats []YearStat, vacations []*Vacation, h Holidays) []string {
	var out []string
	starts := make(map[date.Date]bool, len(stats))
	for _, s := range stats {
		starts[DateOf(s.From)] = true
		used := UsedDaysInYear(vacations, s, h)
		for _, q := range []struct {
			kind       Kind
			used, norm int
		}{
			{KindBase, used.Base, s.Base},
			{KindSeniority, used.Seniority, s.Seniority},
			{KindIrregular, used.Irregular, s.Irregular},
		} {
			if q.used > q.norm {
				out = append(out, fmt.Sprintf("Рабочий год %s: %s — списано %d дн. при норме %d дн.",
					yearRange(s.From, s.To), strings.ToLower(q.kind.Label()), q.used, q.norm))
			}
		}
	}
	if len(stats) == 0 {
		return out
	}

	last := DateOf(stats[len(stats)-1].From)
	for _, v := range vacations {
		if v == nil || !v.Status().IsActive() || !v.Kind().IsAnnual() {
			continue
		}
		start := v.WorkYearStart()
		if starts[start] || start.After(last) {
			continue
		}
		out = append(out, fmt.Sprintf("Отпуск %s–%s привязан к рабочему году с началом %s, которого нет: дни не учтены ни в одном году.",
			v.StartDate().Time().Format("02.01.2006"), v.EndDate().Time().Format("02.01.2006"), start.Time().Format("02.01.2006")))
	}
	return out
}

package vacation

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/calendar"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

// Holidays — нерабочие праздничные дни (ст. 112 ТК РФ). Попадая в период
// отпуска, они не включаются в число его календарных дней (ст. 120 ТК РФ).
// Перенесённые выходные (DayTypeWeekend) сюда не входят и списываются как обычные дни.
type Holidays map[date.Date]struct{}

func HolidaysFrom(exceptions []calendar.Exception) Holidays {
	h := make(Holidays, len(exceptions))
	for _, e := range exceptions {
		if e.DayType().IsHoliday() {
			h[e.Date()] = struct{}{}
		}
	}
	return h
}

func (h Holidays) Has(d date.Date) bool {
	_, ok := h[d]
	return ok
}

// HolidaysRangeEnd — до какой даты нужны праздники, чтобы правильно посчитать
// и дни после at, и уже оформленные отпуска (они могут быть в будущем).
func HolidaysRangeEnd(vacations []*Vacation, at date.Date) date.Date {
	end := at
	for _, v := range vacations {
		if v != nil && v.EndDate().After(end) {
			end = v.EndDate()
		}
	}
	return end
}

// ChargeableDays — сколько дней периода [from, to] списывается из квоты отпуска.
func ChargeableDays(from, to date.Date, h Holidays) int {
	n := 0
	for d := from; !d.After(to); d = d.AddDay() {
		if !h.Has(d) {
			n++
		}
	}
	return n
}

// AllocationOrder — порядок списания видов отпуска внутри одного рабочего года.
var AllocationOrder = []Kind{KindBase, KindIrregular, KindSeniority}

// YearBalance — остаток квоты рабочего года по видам отпуска.
type YearBalance struct {
	Year      YearStat
	Remaining map[Kind]int
}

func (b YearBalance) Total() int {
	total := 0
	for _, n := range b.Remaining {
		total += n
	}
	return total
}

// Allocation — непрерывный кусок отпуска, списанный из квоты одного вида
// одного рабочего года. Праздники внутри куска входят в него, но не в Days.
type Allocation struct {
	Kind          Kind
	WorkYearStart date.Date
	Start         date.Date
	End           date.Date
	Days          int
}

// DateOf переводит начало рабочего года (time.Time) в date.Date.
func DateOf(t time.Time) date.Date {
	return date.New(t.Year(), t.Month(), t.Day())
}

// BuildBalances считает остатки квот по рабочим годам, начавшимся не позже at.
// При capEarned для года, в который попадает at, доступны только дни,
// заработанные к этой дате; иначе — вся годовая норма.
func BuildBalances(
	stats []YearStat,
	vacations []*Vacation,
	h Holidays,
	at date.Date,
	calc WorkYearCalculator,
	capEarned bool,
) ([]YearBalance, error) {
	out := make([]YearBalance, 0, len(stats))

	for _, s := range stats {
		if at.Time().Before(s.From) {
			break
		}

		used := UsedDaysInYear(vacations, s, h)
		remaining := map[Kind]int{
			KindBase:      max(s.Base-used.Base, 0),
			KindIrregular: max(s.Irregular-used.Irregular, 0),
			KindSeniority: max(s.Seniority-used.Seniority, 0),
		}

		if capEarned && at.Time().Before(s.To) {
			usedTotal := used.Base + used.Irregular + used.Seniority
			info, err := calc.CurrentYear([]YearStat{s}, at.Time(), usedTotal)
			if err != nil {
				return nil, fmt.Errorf("earned days for %s: %w", s.From.Format("02.01.2006"), err)
			}
			capLeft := max(int(math.Floor(info.Earned))-usedTotal, 0)
			for _, k := range AllocationOrder {
				take := min(remaining[k], capLeft)
				remaining[k] = take
				capLeft -= take
			}
		}

		out = append(out, YearBalance{Year: s, Remaining: remaining})
	}

	return out, nil
}

// Allocate распределяет дни периодов по квотам: сначала самый старый рабочий год
// с остатком, внутри года — в порядке AllocationOrder. Периоды не должны пересекаться.
//
// При allowOverdraft нехватка дней не ошибка: недостающие дни списываются сверх
// нормы основного отпуска последнего рабочего года из balances; их число — overdraft.
func Allocate(periods []CreationPeriod, balances []YearBalance, h Holidays, allowOverdraft bool) (allocations []Allocation, overdraft int, err error) {
	if len(balances) == 0 {
		return nil, 0, fmt.Errorf("%w: нет рабочего года для списания дней", ErrNoSuitableWorkYear)
	}

	sorted := make([]CreationPeriod, len(periods))
	copy(sorted, periods)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Start.Before(sorted[j].Start) })

	requested := 0
	for _, p := range sorted {
		requested += ChargeableDays(p.Start, p.End, h)
	}
	available := 0
	remaining := make([]map[Kind]int, len(balances))
	for i, b := range balances {
		available += b.Total()
		remaining[i] = make(map[Kind]int, len(b.Remaining))
		for k, n := range b.Remaining {
			remaining[i][k] = n
		}
	}
	if requested > available {
		if !allowOverdraft {
			return nil, 0, fmt.Errorf("%w: к списанию %d дн., доступно %d дн.", ErrExceedsAnnual, requested, available)
		}
		overdraft = requested - available
		// Сверхнормативные дни — в основной отпуск последнего рабочего года.
		remaining[len(remaining)-1][KindBase] += overdraft
	}

	yi, ki := 0, 0
	// bucket находит ближайшую непустую квоту, начиная с текущей.
	bucket := func() bool {
		for yi < len(remaining) {
			if remaining[yi][AllocationOrder[ki]] > 0 {
				return true
			}
			ki++
			if ki == len(AllocationOrder) {
				ki = 0
				yi++
			}
		}
		return false
	}

	out := make([]Allocation, 0, len(sorted))
	for _, p := range sorted {
		var cur *Allocation
		for d := p.Start; !d.After(p.End); d = d.AddDay() {
			holiday := h.Has(d)
			if holiday && cur != nil {
				cur.End = d
				continue
			}
			if !bucket() {
				return nil, 0, fmt.Errorf("%w: к списанию %d дн., доступно %d дн.", ErrExceedsAnnual, requested, available)
			}
			kind, yearStart := AllocationOrder[ki], DateOf(balances[yi].Year.From)
			if cur == nil || cur.Kind != kind || cur.WorkYearStart != yearStart {
				if cur != nil {
					out = append(out, *cur)
				}
				cur = &Allocation{Kind: kind, WorkYearStart: yearStart, Start: d}
			}
			cur.End = d
			if !holiday {
				cur.Days++
				remaining[yi][kind]--
			}
		}
		if cur != nil {
			out = append(out, *cur)
		}
	}

	return out, overdraft, nil
}

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
	// Full — остаток по видам без ограничения заработанным: заработанные дни можно
	// потратить на любой вид, общий лимит года — сумма Remaining.
	Full map[Kind]int
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

		full := make(map[Kind]int, len(remaining))
		for k, n := range remaining {
			full[k] = n
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

		out = append(out, YearBalance{Year: s, Remaining: remaining, Full: full})
	}

	return out, nil
}

// Allocate распределяет дни периодов по квотам: сначала самый старый рабочий год
// с остатком. Период с выбранным видом списывается только из квоты этого вида,
// автоматический (KindAuto) — в порядке AllocationOrder. Периоды с выбранным видом
// распределяются первыми, чтобы автоматические не заняли их дни.
// Периоды не должны пересекаться. Списать больше остатка нельзя — ErrExceedsAnnual.
func Allocate(periods []CreationPeriod, balances []YearBalance, h Holidays) ([]Allocation, error) {
	if len(balances) == 0 {
		return nil, fmt.Errorf("%w: нет рабочего года для списания дней", ErrNoSuitableWorkYear)
	}

	sorted := make([]CreationPeriod, len(periods))
	copy(sorted, periods)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Start.Before(sorted[j].Start) })

	requested := make(map[Kind]int)
	for _, p := range sorted {
		requested[p.Kind] += ChargeableDays(p.Start, p.End, h)
	}
	quotas := newYearQuotas(balances)
	initial := newYearQuotas(balances)

	out := make([]Allocation, 0, len(sorted))
	for _, auto := range []bool{false, true} {
		for _, p := range sorted {
			if (p.Kind == KindAuto) != auto {
				continue
			}
			var cur *Allocation
			for d := p.Start; !d.After(p.End); d = d.AddDay() {
				holiday := h.Has(d)
				if holiday && cur != nil {
					cur.End = d
					continue
				}
				yi, kind, ok := quotas.next(p.Kind)
				if !ok {
					return nil, initial.notEnough(p.Kind, requested)
				}
				yearStart := DateOf(balances[yi].Year.From)
				if cur == nil || cur.Kind != kind || cur.WorkYearStart != yearStart {
					if cur != nil {
						out = append(out, *cur)
					}
					cur = &Allocation{Kind: kind, WorkYearStart: yearStart, Start: d}
				}
				cur.End = d
				if !holiday {
					cur.Days++
					quotas.spend(yi, kind)
				}
			}
			if cur != nil {
				out = append(out, *cur)
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Start.Before(out[j].Start) })

	return out, nil
}

// yearQuotas — изменяемая копия остатков на время распределения: дни по видам
// и общий лимит года (в текущем году он ограничен заработанным).
type yearQuotas struct {
	kinds []map[Kind]int
	left  []int
}

func newYearQuotas(balances []YearBalance) *yearQuotas {
	q := &yearQuotas{kinds: make([]map[Kind]int, len(balances)), left: make([]int, len(balances))}
	for i, b := range balances {
		source := b.Full
		if source == nil {
			source = b.Remaining
		}
		q.kinds[i] = make(map[Kind]int, len(source))
		for k, n := range source {
			q.kinds[i][k] = n
		}
		q.left[i] = b.Total()
	}
	return q
}

// next находит самый старый год и вид, из которых можно списать день.
func (q *yearQuotas) next(k Kind) (year int, kind Kind, ok bool) {
	order := []Kind{k}
	if k == KindAuto {
		order = AllocationOrder
	}
	for yi := range q.kinds {
		if q.left[yi] <= 0 {
			continue
		}
		for _, kind := range order {
			if q.kinds[yi][kind] > 0 {
				return yi, kind, true
			}
		}
	}
	return 0, "", false
}

func (q *yearQuotas) spend(year int, k Kind) {
	q.kinds[year][k]--
	q.left[year]--
}

// notEnough — ошибка нехватки дней; q — квоты до начала распределения.
func (q *yearQuotas) notEnough(k Kind, requested map[Kind]int) error {
	if k == KindAuto {
		available := 0
		for _, n := range q.left {
			available += n
		}
		for kind, n := range requested {
			if kind != KindAuto {
				available -= n
			}
		}
		return fmt.Errorf("%w: к списанию %d дн., доступно %d дн.", ErrExceedsAnnual, requested[KindAuto], max(available, 0))
	}
	available := 0
	for yi := range q.kinds {
		available += min(q.kinds[yi][k], q.left[yi])
	}
	return fmt.Errorf("%w: вид отпуска «%s»: к списанию %d дн., доступно %d дн.", ErrExceedsAnnual, k.Label(), requested[k], available)
}

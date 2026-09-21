package vacation

type Reason string

const (
	ReasonBase      Reason = "base"      // основной отпуск
	ReasonSeniority Reason = "seniority" // выслуга лет
	ReasonIrregular Reason = "irregular" // ненормированный слуюебный день
)

type AllowanceItem struct {
	Reason Reason
	Days   int
}

type Allowance struct {
	Items []AllowanceItem
}

func (a Allowance) Total() int {
	total := 0
	for _, it := range a.Items {
		total += it.Days
	}
	return total
}

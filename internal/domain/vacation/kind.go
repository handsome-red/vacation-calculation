package vacation

import "fmt"

type Kind string

const (
	KindBase      Kind = "base"      // основной отпуск
	KindSeniority Kind = "seniority" // выслуга лет
	KindIrregular Kind = "irregular" // ненормированный слуюебный день
)

func NewKind(s string) (Kind, error) {
	k := Kind(s)
	switch k {
	case KindBase, KindSeniority, KindIrregular:
		return k, nil
	default:
		return "", fmt.Errorf("Unknown kind")
	}
}

type AllowanceItem struct {
	Kind Kind
	Days int
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

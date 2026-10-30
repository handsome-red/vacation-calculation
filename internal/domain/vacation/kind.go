package vacation

import "fmt"

type Kind string

const (
	KindBase      Kind = "base"      // основной отпуск
	KindSeniority Kind = "seniority" // выслуга лет
	KindIrregular Kind = "irregular" // ненормированный слуюебный день
	// KindUnpaid — административный отпуск без сохранения заработной платы.
	// Не расходует квоту; сверх 14 дней за рабочий год сдвигает рабочий год.
	KindUnpaid Kind = "unpaid"
)

func NewKind(s string) (Kind, error) {
	k := Kind(s)
	switch k {
	case KindBase, KindSeniority, KindIrregular, KindUnpaid:
		return k, nil
	default:
		return "", fmt.Errorf("Unknown kind")
	}
}

// IsAnnual — ежегодный оплачиваемый отпуск (списывается из квоты рабочего года).
func (k Kind) IsAnnual() bool {
	return k == KindBase || k == KindIrregular || k == KindSeniority
}

var kindLabels = map[Kind]string{
	KindBase:      "Основной",
	KindIrregular: "За ненормированный день",
	KindSeniority: "За выслугу лет",
	KindUnpaid:    "Без сохранения заработной платы",
}

func (k Kind) Label() string {
	if l, ok := kindLabels[k]; ok {
		return l
	}
	return string(k)
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

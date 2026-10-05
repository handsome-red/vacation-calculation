package vacation

import "time"

type ShiftKind string

const (
	ShiftKindUnpaidLong    ShiftKind = "unpaid_long"
	ShiftKindParentalLeave ShiftKind = "parental_leave"
	ShiftKindAbsenteeism   ShiftKind = "absenteeism"
)

var shiftMap = map[ShiftKind]string{
	ShiftKindUnpaidLong:	 	"Неоплачиваемый отпуск",
	ShiftKindParentalLeave: 	"Уход за ребенком",
	ShiftKindAbsenteeism: 		"Прогул",
}

type Shift struct {
	Kind ShiftKind
	From time.Time
	To   time.Time
}
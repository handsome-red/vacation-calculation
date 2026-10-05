package vacation

import (
	"errors"
	"fmt"
	"time"
)

type ShiftKind string

const (
	ShiftKindUnpaid        ShiftKind = "unpaid"
	ShiftKindParentalLeave ShiftKind = "parental_leave"
	ShiftKindAbsenteeism   ShiftKind = "absenteeism"
)

var shiftMap = map[ShiftKind]string{
	ShiftKindUnpaid:        "Неоплачиваемый отпуск",
	ShiftKindParentalLeave: "Уход за ребенком",
	ShiftKindAbsenteeism:   "Прогул",
}

type Shift struct {
	Kind ShiftKind
	From time.Time // включительно
	To   time.Time // включительно
}

// Перенести Error
func (s Shift) Validate() error {
	if s.To.Before(s.From) {
		return fmt.Errorf("shift: To (%s) раньше From (%s)", s.To, s.From)
	}
	if s.Kind == "" {
		return errors.New("shift: пустой Kind")
	}
	if _, ok := shiftMap[s.Kind]; !ok {
		return fmt.Errorf("shift: неизвестный Kind: %s", s.Kind)
	}
	return nil
}

func (s Shift) Days() int {
	if s.To.Before(s.From) {
		return 0
	}
	return int(s.To.Sub(s.From).Hours()/24) + 1
}

func (s Shift) Contains(t time.Time) bool {
	return !t.Before(s.From) && !t.After(s.To)
}

func shiftDaysForYear(yearStart, yearEnd time.Time, shifts []Shift) int {
	lastDay := yearEnd.AddDate(0, 0, -1)
	var full, unpaid int
	for _, s := range shifts {
		oStart := maxTime(s.From, yearStart)
		oEnd := minTime(s.To, lastDay)
		if oEnd.Before(oStart) {
			continue
		}
		d := int(oEnd.Sub(oStart).Hours()/24) + 1
		switch s.Kind {
		case ShiftKindUnpaid:
			unpaid += d
		case ShiftKindParentalLeave, ShiftKindAbsenteeism:
			full += d
		}
	}

	if unpaid > 14 {
		full += unpaid - 14
	}

	return full
}

func maxTime(time, yearStart time.Time) time.Time {
	if time.After(yearStart) {
		return time
	}
	return yearStart
}

func minTime(time, lastDay time.Time) time.Time {
	if time.Before(lastDay) {
		return time
	}
	return lastDay
}

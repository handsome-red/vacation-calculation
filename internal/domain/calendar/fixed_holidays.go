package calendar

import (
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

// FixedHolidays — нерабочие праздничные дни года с постоянными датами
// (ст. 112 ТК РФ и День Республики Татарстан). Ураза-байрам и Курбан-байрам
// зависят от лунного календаря, а переносы выходных — от постановления
// правительства: их администратор добавляет вручную.
func FixedHolidays(year int) []Exception {
	type h struct {
		m    time.Month
		d    int
		name string
	}
	list := []h{
		{time.January, 1, "Новогодние каникулы"},
		{time.January, 2, "Новогодние каникулы"},
		{time.January, 3, "Новогодние каникулы"},
		{time.January, 4, "Новогодние каникулы"},
		{time.January, 5, "Новогодние каникулы"},
		{time.January, 6, "Новогодние каникулы"},
		{time.January, 7, "Рождество Христово"},
		{time.January, 8, "Новогодние каникулы"},
		{time.February, 23, "День защитника Отечества"},
		{time.March, 8, "Международный женский день"},
		{time.May, 1, "Праздник Весны и Труда"},
		{time.May, 9, "День Победы"},
		{time.June, 12, "День России"},
		{time.August, 30, "День Республики Татарстан"},
		{time.November, 4, "День народного единства"},
	}

	out := make([]Exception, 0, len(list))
	for _, x := range list {
		e, _ := NewException(date.New(year, x.m, x.d), DayTypeHoliday, x.name)
		out = append(out, e)
	}
	return out
}

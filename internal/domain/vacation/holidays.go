package vacation

import "time"

// TODO: Добавить новогодние каникулы и мусульманские праздники
type Holiday struct {
	Name  string
	Month time.Month
	Day   int
}

var Holidays = []Holiday{
	{Name: "Новый год", Month: time.January, Day: 1},
	{Name: "Рождество Христово", Month: time.January, Day: 7},
	{Name: "День защитника Отечества", Month: time.February, Day: 23},
	{Name: "Международный женский день", Month: time.March, Day: 8},
	{Name: "Праздник Весны и Труда", Month: time.May, Day: 1},
	{Name: "День Победы", Month: time.May, Day: 9},
	{Name: "День России", Month: time.June, Day: 12},
	{Name: "День народного единства", Month: time.November, Day: 4},
	{Name: "День Республики Татарстан", Month: time.August, Day: 30},
}

func IsHoliday(t time.Time) (string, bool) {
	for _, h := range Holidays {
		if h.Month == t.Month() && h.Day == t.Day() {
			return h.Name, true
		}
	}
	return "", false
}

func AllHolidays() []Holiday {
	return Holidays
}

package calendar

type DayType bool

const (
	DayTypeHoliday DayType = true
	DayTypeWorking DayType = false
)

func (t DayType) Valid() bool {
	return t == DayTypeWorking || t == DayTypeHoliday
}

package calendar

type DayType string

const (
	DayTypeWorking      DayType = "working"
    DayTypeWeekend      DayType = "weekend"
    DayTypeHoliday      DayType = "holiday"
)

func (d DayType) IsWeekend() bool {
	return d == DayTypeWeekend
}
func (d DayType) IsHoliday() bool {
	return d == DayTypeHoliday
}

func (t DayType) Valid() bool {
    switch t {
    case DayTypeWorking, DayTypeWeekend, DayTypeHoliday:
        return true
    }
    return false
}
package user

import "fmt"

type Experience struct {
	Years     int
	Months    int
	Days      int
	TotalDays int
}

func (e Experience) String() string {
	return fmt.Sprintf("%dг./л. %dм. %dд. (всего: %dд.)", e.Years, e.Months, e.Days, e.TotalDays)
}

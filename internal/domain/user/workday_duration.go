package user

import "fmt"

type WorkdayDuration int

func (w WorkdayDuration) Int() int {
return int(w)
}

func NewWorkdayDuration(v int) (WorkdayDuration, error) {
w := WorkdayDuration(v)
if !w.isValid() {
return 0, ErrWorkdayDurationInvalid
}
return w, nil
}

func (w WorkdayDuration) isValid() bool {
if w < 1 || w > 8 {
return false
}

return true
}

// String returns the workday duration as a human-readable string.
func (w WorkdayDuration) String() string {
return fmt.Sprintf("%d ч.", int(w))
}

package create_shift

import "time"

type Command struct {
	UserID string
	Kind   string // "unpaid_long" | "parental_leave" | "absenteeism"
	From   time.Time
	To     time.Time
}

type Result struct {
}

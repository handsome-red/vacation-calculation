package user

import "strings"

// import "fmt"

// Status - статус пользователя
type Status string

const (
	StatusUnknown Status = ""
	StatusActive  Status = "active"
	StatusBlocked Status = "blocked"
	StatusFired   Status = "fired"
	StatusDeleted Status = "deleted"
)

func NewStatus(s string) (Status, error) {
	st := Status(strings.ToLower(strings.TrimSpace(s)))
	if !st.isValid() {
		return StatusUnknown, ErrStatusInvalid
	}
	return st, nil
}

func (s Status) isValid() bool {
	switch s {
	case StatusActive, StatusBlocked, StatusFired, StatusDeleted:
		return true
	default:
		return false
	}
}

func (s Status) String() string { return string(s) }

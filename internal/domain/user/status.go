package user

import "strings"

// Status - статус пользователя
type Status string

const (
	StatusUnknown Status = ""
	StatusActive  Status = "active"
	StatusBlocked Status = "blocked"
	StatusFired   Status = "fired"
)

// Удален не отображается
// TODO
var statusMap = map[Status]string{
	StatusActive:  "Работает",
	StatusBlocked: "Заблокирован", // декрет, административный
	StatusFired:   "Уволен",
}

func NewStatus(s string) (Status, error) {
	st := Status(strings.ToLower(strings.TrimSpace(s)))
	if !st.isValid() {
		return StatusUnknown, ErrStatusInvalid
	}
	return st, nil
}

func (s Status) IsActive() bool {
	return s == StatusActive
}

func (s Status) isValid() bool {
	switch s {
	case StatusActive, StatusBlocked, StatusFired:
		return true
	default:
		return false
	}
}

func (s Status) Label() string { return statusMap[s] }

// String returns the database-safe value for serialization.
func (s Status) String() string {
	switch s {
	case StatusActive, StatusBlocked, StatusFired, StatusUnknown:
		return string(s)
	default:
		return ""
	}
}

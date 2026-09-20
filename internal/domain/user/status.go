package user

// import "fmt"

// Status - статус пользователя
type Status int

// TODO Какие еще могут быть статусы?
const (
	StatusWork  Status = iota // Неизвестный
	StatusFired               // Работает
)

// NewStatus - создаёт Status из int с валидацией
func NewStatus(v int) (Status, error) {
	s := Status(v)
	// if !s.IsValid() {
		// return StatusUnknown, fmt.Errorf("invalid status: %d", v)
	// }
	return s, nil
}

// IsActive - активен ли пользователь
// func (s Status) IsActive() bool { return s == StatusActive }

// String - строковое представление
func (s Status) String() string {
	switch s {
	// case StatusActive:
	// 	return "active"
	// case StatusBlocked:
	// 	return "blocked"
	// case StatusDeleted:
	// 	return "deleted"
	default:
		return "unknown"
	}
}

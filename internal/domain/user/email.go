package user

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

type Email struct {
	value string
}

// NewEmail создает новый Email с валидацией.
// Возвращает ошибку, если email некорректен.
func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if err := validateEmail(email); err != nil {
		return Email{}, err
	}

	return Email{value: email}, nil
}

// Value возвращает строковое представление Email.
func (e Email) Value() string {
	return e.value
}

// String реализует интерфейс fmt.Stringer
func (e Email) String() string {
	return e.value
}

// validateEmail проверяет корректность email.
func validateEmail(email string) error {
	if email == "" {
		return ErrEmailRequired
	}

	if !emailRegex.MatchString(email) {
		return ErrInvalidEmailFormat
	}

	return nil
}

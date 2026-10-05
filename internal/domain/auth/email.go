package auth

import "strings"

type Email struct {
	value string
}

func NewEmail(s string) (Email, error) {
	// TODO
	return Email{value: s}, nil
}

// TODO
func(e Email)isValid(s string) bool {
	return strings.Contains(s, "@")
}

func(e Email)Value() string {
	return e.value
}
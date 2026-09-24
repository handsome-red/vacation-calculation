package user

import (
	"fmt"

	"github.com/google/uuid"
)

type UserID struct {
	value uuid.UUID
}

func (id UserID) IsZero() bool {
	return id.value == uuid.Nil
}

// GenerateUserID генерирует новый случайный ID
func GenerateUserID() UserID {
	return UserID{value: uuid.New()}
}

// NewUserID создает новый UserID.
func NewUserID(id uuid.UUID) (UserID, error) {
	if id == uuid.Nil {
		return UserID{}, ErrIDEmpty
	}
	return UserID{value: id}, nil
}

// String возвращает строковое представление ID.
func (id UserID) String() string {
	return id.value.String()
}

// Value возвращает значение ID.
func (id UserID) UUID() uuid.UUID {
	return id.value
}

// ParseUserID парсит ID.
func ParseUserID(s string) (UserID, error) {
	uuid, err := uuid.Parse(s)
	if err != nil {
		return UserID{}, fmt.Errorf("parse user id: %w", err)
	}

	return UserID{value: uuid}, nil
}

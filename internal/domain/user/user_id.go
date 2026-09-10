package user

import "github.com/google/uuid"

type UserID struct {
	value uuid.UUID
}

// GenerateUserID генерирует новый случайный ID
func GenerateUserID() UserID {
	return UserID{value: uuid.New()}
}

// NewUserID создает новый UserID.
func NewUserID(id uuid.UUID) (UserID, error) {
	if id == uuid.Nil {
		return UserID{}, ErrIDIsEmpty
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

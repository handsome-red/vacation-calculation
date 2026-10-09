package auth

import "github.com/google/uuid"

type UserID struct {
	value uuid.UUID
}

func ParseUserID(s string) (UserID, error) {
	value, err := uuid.Parse(s)
	if err != nil {
		return UserID{}, ErrIDRequired
	}
	return UserID{
		value: value,
	}, nil
}

func (u UserID) String() string {
	return u.value.String()
}

func (u UserID) IsZero() bool {
	return u.value == uuid.Nil
}

package auth

import "github.com/google/uuid"

type UserID struct {
	value uuid.UUID
}

func ParseUserID(s string) (UserID, error) {
	value, err := uuid.Parse(s)
	if err != nil {
		return  UserID{}, ErrIDRequired
	}
	return UserID{
		value: value,
	}, nil
}

func(u UserID) isValid(s string) bool {
	userID, err := uuid.Parse(s)
	if err != nil {
		return false
	}
	return userID != uuid.Nil
}

func(u UserID) IsZero() bool {
	return u.value != uuid.Nil
}
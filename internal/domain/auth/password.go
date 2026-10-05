package auth

import "strings"

type PasswordHash struct {
	value string
}

func NewPasswordHashUnsafe(hash string) PasswordHash {
    return PasswordHash{value: hash}
}

func NewPasswordHash(hash string) (PasswordHash, error) {
    hash = strings.TrimSpace(hash)
    if hash == "" {
        return PasswordHash{}, ErrPasswordRequired
    }
    return PasswordHash{value: hash}, nil
}

func (h PasswordHash) Value() string {
    return h.value
}

func (h PasswordHash) IsEmpty() bool {
    return h.value == ""
}

func (h PasswordHash) Equal(other PasswordHash) bool {
    return h.value == other.value
}
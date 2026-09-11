package hasher

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < bcrypt.MinCost {
		cost = bcrypt.MinCost
	}
	if cost > bcrypt.MaxCost {
		cost = bcrypt.MaxCost
	}

	return &BcryptHasher{cost: cost}
}

var _ ports.PasswordHasher = (*BcryptHasher)(nil)

func (h *BcryptHasher) HashPassword(ctx context.Context, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}

	return string(hash), nil
}

// Verify проверяет пароль
func (h *BcryptHasher) Verify(ctx context.Context, hashedPassword, plainPassword string) (bool, error) {
	if hashedPassword == "" {
		return false, errors.New("hashed password is empty")
	}
	if plainPassword == "" {
		return false, errors.New("plain password is empty")
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(plainPassword),
	)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	return false, fmt.Errorf("verifying password: %w", err)
}

// NeedsRehash проверяет, нужно ли перехешировать
func (h *BcryptHasher) NeedsRehash(ctx context.Context, hashedPassword string) bool {
	if hashedPassword == "" {
		return true // Пустой хеш точно нужно пересоздать
	}

	hashCost, err := bcrypt.Cost([]byte(hashedPassword))
	if err != nil {
		// Хеш повреждён — нужно пересоздать
		return true
	}

	return hashCost != h.cost
}

package ports

import (
	"context"
)

type PasswordHasher interface {
	// Hash хеширует пароль
	HashPassword(ctx context.Context, password string) (string, error)

	// Verify проверяет пароль
	Verify(ctx context.Context, hashedPassword, plainPassword string) (bool, error)

	// NeedsRehash проверяет, нужно ли перехешировать
	NeedsRehash(ctx context.Context, hashedPassword string) bool
}

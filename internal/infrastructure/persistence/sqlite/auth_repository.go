package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/jmoiron/sqlx"
)

type authRepository struct {
	db *sqlx.DB
}

var _ ports.AuthRepository = (*authRepository)(nil)

func NewAuthRepository(db *sqlx.DB) *authRepository {
	return &authRepository{
		db: db,
	}
}

func (r authRepository) FindByEmail(ctx context.Context, email auth.Email) (*auth.AuthUser, error) {
	const q = `
		SELECT id, email, password_hash, role, status
		FROM auth_users
		WHERE email = ?
	`

	var result authRow
	if err := r.db.GetContext(ctx, &result, q, email.Value()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("find auth user by email: %w", err)
	}

	return result.toDomain()
}

func (r authRepository) Save(ctx context.Context, u *auth.AuthUser) error {
	const q = `
		INSERT INTO auth_users(
			id, email, password_hash, role, status
		)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			id					= EXCLUDED.id,
			email				= EXCLUDED.email,
			password_hash		= EXCLUDED.password_hash,
			role				= EXCLUDED.role,
			status				= EXCLUDED.status,
	`

	_, err := r.db.ExecContext(ctx, q,
		u.ID(),
		u.Email(),
		u.PasswordHash(),
		u.Role(),
		u.Status(),
	)
	if err != nil {
		return fmt.Errorf("saving auth user: %w", err)
	}

	return nil
}

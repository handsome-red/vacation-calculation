package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/jmoiron/sqlx"
)

type sessionRepository struct {
	db *sqlx.DB
}

var _ ports.SessionRepository = (*sessionRepository)(nil)

func NewSessionRepository(db *sqlx.DB) *sessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, tokenHash string, userID user.UserID, expires time.Time) error {
	const q = `INSERT INTO sessions (token_hash, user_id, expires_at) VALUES (?, ?, ?)`
	if _, err := r.db.ExecContext(ctx, q, tokenHash, userID.String(), expires.Unix()); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *sessionRepository) Find(ctx context.Context, tokenHash string, now time.Time) (user.UserID, bool, error) {
	var raw string
	const q = `SELECT user_id FROM sessions WHERE token_hash = ? AND expires_at > ?`
	if err := r.db.GetContext(ctx, &raw, q, tokenHash, now.Unix()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user.UserID{}, false, nil
		}
		return user.UserID{}, false, fmt.Errorf("find session: %w", err)
	}
	id, err := user.ParseUserID(raw)
	if err != nil {
		return user.UserID{}, false, fmt.Errorf("session user id %q: %w", raw, err)
	}
	return id, true, nil
}

func (r *sessionRepository) Delete(ctx context.Context, tokenHash string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = ?`, tokenHash); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (r *sessionRepository) DeleteExpired(ctx context.Context, now time.Time) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= ?`, now.Unix()); err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}
	return nil
}

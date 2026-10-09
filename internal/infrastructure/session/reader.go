package session

import (
	"context"
	"errors"
	"fmt"

	"github.com/handsome-red/vacation-calculation/internal/application/queries/auth/get_current_user"
	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

type reader struct {
	q *get_current_user.Handler
}

func NewReader(q *get_current_user.Handler) *reader {
	return &reader{
		q: q,
	}
}

func (r *reader) UserBySession(ctx context.Context, sessionID string) (*auth.AuthUser, error) {
	user, err := r.q.Handle(ctx, get_current_user.Query{SessionID: sessionID})
	if err != nil {
		if errors.Is(err, auth.ErrSessionNotFound) {
			return nil, ports.ErrSessionNotFound
		}
		return nil, fmt.Errorf("session reader: %w", err)
	}
	return &user.User, nil
}

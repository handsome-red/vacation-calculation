package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

const sessionCookieName = "session_id"

type contextKey struct{ name string }

var authUserKey = contextKey{
	name: "authUser",
}

type AuthMiddleware struct {
	sessions  ports.SessionRepository
	loginPath string
	logger    ports.Logger
}

func NewAuthMiddleware(
	sessions ports.SessionRepository,
	loginPath string,
	logger ports.Logger,
) *AuthMiddleware {
	return &AuthMiddleware{
		sessions:  sessions,
		loginPath: loginPath,
		logger:    logger,
	}
}

func (m *AuthMiddleware) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := m.auntificate(r)
		if !ok {
			m.unauthorized(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), authUserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) auntificate(r *http.Request) (*auth.AuthUser, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return nil, false
	}

	user, err := m.sessions.UserBySession(r.Context(), cookie.Value)
	switch {
	case err == nil:
		return user, true
	case errors.Is(err, ports.ErrSessionNotFound):
		return nil, false
	default:
		m.logger.Info(r.Context(), "auth: resolve session",
			"error", err.Error(),
			"path", r.URL.Path,
		)
		return nil, false
	}
}

func (m *AuthMiddleware) unauthorized(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, m.loginPath, http.StatusSeeOther)
}

func UserFromContext(ctx context.Context) (*auth.AuthUser, bool) {
	u, ok := ctx.Value(authUserKey).(*auth.AuthUser)
	return u, ok
}

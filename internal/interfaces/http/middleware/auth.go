package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

const sessionCookieName = "session_id"
const sessionLifetime = 24 * time.Hour

type contextKey string

const userCtxKey contextKey = "user"

type session struct {
	userID  user.UserID
	expires time.Time
}

type AuthMiddleware struct {
	users    ports.UserRepository
	mu       sync.Mutex
	sessions map[string]session
}

func NewAuthMiddleware(users ports.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{users: users, sessions: make(map[string]session)}
}

func (m *AuthMiddleware) Start(w http.ResponseWriter, r *http.Request, id user.UserID) error {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return err
	}
	token := hex.EncodeToString(key)
	m.mu.Lock()
	m.sessions[token] = session{userID: id, expires: time.Now().Add(sessionLifetime)}
	m.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: token, Path: "/", HttpOnly: true,
		Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode, MaxAge: int(sessionLifetime.Seconds())})
	return nil
}

func (m *AuthMiddleware) End(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		m.mu.Lock()
		delete(m.sessions, cookie.Value)
		m.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Path: "/", HttpOnly: true,
		Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode, MaxAge: -1})
}

func (m *AuthMiddleware) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err == nil {
			m.mu.Lock()
			s, found := m.sessions[cookie.Value]
			if found && time.Now().After(s.expires) {
				delete(m.sessions, cookie.Value)
				found = false
			}
			m.mu.Unlock()
			if found {
				u, err := m.users.FindByID(r.Context(), s.userID)
				if err == nil && u.Status().IsActive() {
					next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userCtxKey, u)))
					return
				}
			}
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.Error(w, "Необходим вход", http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
}

func UserFromContext(ctx context.Context) (*user.User, bool) {
	u, ok := ctx.Value(userCtxKey).(*user.User)
	return u, ok
}

func CallerFromContext(ctx context.Context) (auth.Caller, bool) {
	u, ok := UserFromContext(ctx)
	if !ok || u == nil {
		return auth.Caller{}, false
	}
	return auth.Caller{UserID: u.ID(), IsAdmin: u.IsAdmin()}, true
}

// RequireAdmin разрешает доступ только суперадминистратору после Require.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || u == nil {
			http.Error(w, "Необходим вход", http.StatusUnauthorized)
			return
		}
		if !u.IsAdmin() {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

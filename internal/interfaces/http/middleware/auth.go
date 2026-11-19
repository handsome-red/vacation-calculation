package middleware

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

const sessionCookieName = "session_id"
const sessionLifetime = 24 * time.Hour

type contextKey string

const userCtxKey contextKey = "user"

// AuthMiddleware ведёт сессии входа. Сессии хранятся в базе и переживают перезапуск.
type AuthMiddleware struct {
	users    ports.UserRepository
	sessions ports.SessionRepository
}

func NewAuthMiddleware(users ports.UserRepository, sessions ports.SessionRepository) *AuthMiddleware {
	return &AuthMiddleware{users: users, sessions: sessions}
}

// hashToken — ключ сессии в базе: SHA-256 токена из куки.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (m *AuthMiddleware) Start(w http.ResponseWriter, r *http.Request, id user.UserID) error {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return err
	}
	token := hex.EncodeToString(key)
	now := time.Now()
	// Истекшие сессии убираются при каждом входе: входов мало, отдельная задача не нужна.
	if err := m.sessions.DeleteExpired(r.Context(), now); err != nil {
		return err
	}
	if err := m.sessions.Create(r.Context(), hashToken(token), id, now.Add(sessionLifetime)); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: token, Path: "/", HttpOnly: true,
		Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode, MaxAge: int(sessionLifetime.Seconds())})
	return nil
}

func (m *AuthMiddleware) End(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		_ = m.sessions.Delete(r.Context(), hashToken(cookie.Value))
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Path: "/", HttpOnly: true,
		Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode, MaxAge: -1})
}

func (m *AuthMiddleware) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err == nil {
			userID, found, err := m.sessions.Find(r.Context(), hashToken(cookie.Value), time.Now())
			if err != nil {
				http.Error(w, "Ошибка проверки сессии", http.StatusInternalServerError)
				return
			}
			if found {
				u, err := m.users.FindByID(r.Context(), userID)
				if err == nil && u.Status().IsActive() {
					next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), u)))
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

// WithUser — контекст запроса вошедшего пользователя u (как после Require).
func WithUser(ctx context.Context, u *user.User) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
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
	return auth.CallerOf(u), true
}

// RequireSelfOrManager пропускает пользователя, чей ID указан в параметре маршрута
// param, и тех, кто им управляет: администратора и руководителя его территориального
// отдела. Применяется после Require.
func (m *AuthMiddleware) RequireSelfOrManager(param string) func(http.Handler) http.Handler {
	return m.requireTarget(param, true)
}

// RequireManagerOf пропускает только тех, кто управляет сотрудником из параметра
// маршрута param: администратора и руководителя его территориального отдела.
func (m *AuthMiddleware) RequireManagerOf(param string) func(http.Handler) http.Handler {
	return m.requireTarget(param, false)
}

func (m *AuthMiddleware) requireTarget(param string, allowSelf bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFromContext(r.Context())
			if !ok || u == nil {
				http.Error(w, "Необходим вход", http.StatusUnauthorized)
				return
			}
			raw := chi.URLParam(r, param)
			if allowSelf && strings.EqualFold(raw, u.ID().String()) {
				next.ServeHTTP(w, r)
				return
			}
			if u.IsAdmin() {
				next.ServeHTTP(w, r)
				return
			}
			if !u.IsManager() {
				http.Error(w, "Доступ запрещён", http.StatusForbidden)
				return
			}
			id, err := user.ParseUserID(raw)
			if err != nil {
				http.Error(w, "Сотрудник не найден", http.StatusNotFound)
				return
			}
			target, err := m.users.FindByID(r.Context(), id)
			if err != nil {
				http.Error(w, "Сотрудник не найден", http.StatusNotFound)
				return
			}
			if !u.Manages(target) {
				http.Error(w, "Доступ запрещён", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireStaff пропускает администратора и руководителей. Доступные им сотрудники
// ограничиваются в обработчиках правами из CallerFromContext.
func RequireStaff(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || u == nil {
			http.Error(w, "Необходим вход", http.StatusUnauthorized)
			return
		}
		if !u.IsStaff() {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
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

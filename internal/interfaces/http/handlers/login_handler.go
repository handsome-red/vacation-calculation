package handlers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/auth/login"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/middleware"
)

type LoginHandler struct {
	login     *login.Handler
	sessions  *middleware.AuthMiddleware
	limiter   *middleware.LoginLimiter
	templates *Templates
}

func NewLoginHandler(login *login.Handler, sessions *middleware.AuthMiddleware, limiter *middleware.LoginLimiter, templates *Templates) *LoginHandler {
	return &LoginHandler{login: login, sessions: sessions, limiter: limiter, templates: templates}
}

func (h *LoginHandler) Form(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if err := h.templates.Render(w, "login.html", struct{ Error string }{}); err != nil {
		http.Error(w, "Ошибка отображения формы", http.StatusInternalServerError)
	}
}

func (h *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !sameOrigin(r) {
		http.Error(w, "Недопустимый источник", http.StatusForbidden)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}
	email, ip := r.PostFormValue("email"), middleware.ClientIP(r)
	if wait, blocked := h.limiter.Blocked(ip, email); blocked {
		minutes := int(math.Ceil(wait.Minutes()))
		w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(wait.Seconds()))))
		w.WriteHeader(http.StatusTooManyRequests)
		_ = h.templates.Render(w, "login.html", struct{ Error string }{
			Error: fmt.Sprintf("Слишком много неудачных попыток входа. Повторите через %d мин.", minutes)})
		return
	}
	result, err := h.login.Handle(r.Context(), login.Command{Email: email, Password: r.PostFormValue("password")})
	if errors.Is(err, login.ErrInvalidCredentials) {
		h.limiter.Fail(ip, email)
		w.WriteHeader(http.StatusUnauthorized)
		_ = h.templates.Render(w, "login.html", struct{ Error string }{Error: "Неверная почта или пароль"})
		return
	}
	if err != nil {
		http.Error(w, "Ошибка входа", http.StatusInternalServerError)
		return
	}
	id, err := user.ParseUserID(result.UserID)
	if err != nil {
		http.Error(w, "Ошибка входа", http.StatusInternalServerError)
		return
	}
	if err := h.sessions.Start(w, r, id); err != nil {
		http.Error(w, "Ошибка входа", http.StatusInternalServerError)
		return
	}
	h.limiter.Success(email)
	if result.IsStaff {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/users/"+result.UserID, http.StatusSeeOther)
}

func (h *LoginHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		http.Error(w, "Недопустимый источник", http.StatusForbidden)
		return
	}
	h.sessions.End(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func sameOrigin(r *http.Request) bool { return middleware.IsSameOrigin(r) }

package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/auth/login"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/middleware"
)

type LoginHandler struct {
	login     *login.Handler
	sessions  *middleware.AuthMiddleware
	templates *Templates
}

func NewLoginHandler(login *login.Handler, sessions *middleware.AuthMiddleware, templates *Templates) *LoginHandler {
	return &LoginHandler{login: login, sessions: sessions, templates: templates}
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
	result, err := h.login.Handle(r.Context(), login.Command{Email: r.PostFormValue("email"), Password: r.PostFormValue("password")})
	if errors.Is(err, login.ErrInvalidCredentials) {
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
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *LoginHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		http.Error(w, "Недопустимый источник", http.StatusForbidden)
		return
	}
	h.sessions.End(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	return origin == "" || strings.EqualFold(origin, "http://"+r.Host) || strings.EqualFold(origin, "https://"+r.Host)
}

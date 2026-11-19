package middleware

import (
	"net/http"
	"strings"
)

// IsSameOrigin — запрос отправлен со страницы этого же сайта. Браузеры указывают
// источник в Origin (и в Sec-Fetch-Site); запросы без этих заголовков — не из
// браузера (curl, тесты), CSRF им не грозит.
func IsSameOrigin(r *http.Request) bool {
	if site := r.Header.Get("Sec-Fetch-Site"); site == "cross-site" || site == "same-site" {
		return false
	}
	origin := r.Header.Get("Origin")
	return origin == "" || strings.EqualFold(origin, "http://"+r.Host) || strings.EqualFold(origin, "https://"+r.Host)
}

// SameOrigin отклоняет изменяющие запросы с чужих сайтов — защита от CSRF для всех
// форм и API сразу. Сессионная кука с SameSite=Lax и так не уходит с чужих POST,
// проверка закрывает остальные случаи (поддомены, старые браузеры).
func SameOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
		default:
			if !IsSameOrigin(r) {
				http.Error(w, "Недопустимый источник запроса", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

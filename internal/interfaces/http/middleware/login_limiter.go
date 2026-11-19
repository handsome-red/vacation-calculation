package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Пределы неудачных входов за окно: на e-mail — против подбора пароля к учётной
// записи, на адрес — против перебора многих учётных записей с одного адреса.
const (
	loginWindow     = 10 * time.Minute
	loginEmailLimit = 5
	loginIPLimit    = 20
	// loginMaxKeys — при стольких ключах устаревшие записи вычищаются целиком.
	loginMaxKeys = 10000
)

// LoginLimiter считает неудачные попытки входа в памяти. После перезапуска счётчики
// обнуляются — для ограничения подбора этого достаточно.
type LoginLimiter struct {
	mu    sync.Mutex
	fails map[string][]time.Time
	now   func() time.Time
}

func NewLoginLimiter() *LoginLimiter {
	return &LoginLimiter{fails: make(map[string][]time.Time), now: time.Now}
}

// ClientIP — адрес клиента из соединения. Приложение принимает соединения напрямую,
// поэтому заголовкам X-Forwarded-For не доверяем.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func emailKey(email string) string { return "email:" + strings.ToLower(strings.TrimSpace(email)) }
func ipKey(ip string) string       { return "ip:" + ip }

// Blocked сообщает, исчерпан ли предел для адреса или e-mail, и через сколько можно
// повторить попытку.
func (l *LoginLimiter) Blocked(ip, email string) (time.Duration, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	var wait time.Duration
	for _, c := range []struct {
		key   string
		limit int
	}{{emailKey(email), loginEmailLimit}, {ipKey(ip), loginIPLimit}} {
		times := l.recent(c.key, now)
		if len(times) >= c.limit {
			// Попытка станет доступна, когда из окна выйдет самая ранняя из последних limit.
			if d := times[len(times)-c.limit].Add(loginWindow).Sub(now); d > wait {
				wait = d
			}
		}
	}
	return wait, wait > 0
}

// Fail учитывает неудачную попытку входа.
func (l *LoginLimiter) Fail(ip, email string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	if len(l.fails) >= loginMaxKeys {
		for key := range l.fails {
			l.recent(key, now)
		}
	}
	for _, key := range []string{emailKey(email), ipKey(ip)} {
		l.fails[key] = append(l.recent(key, now), now)
	}
}

// Success сбрасывает счётчик e-mail после успешного входа. Счётчик адреса остаётся:
// иначе вход в свою учётную запись открывал бы перебор чужих.
func (l *LoginLimiter) Success(email string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.fails, emailKey(email))
}

// recent оставляет у ключа только попытки внутри окна. Вызывается под l.mu.
func (l *LoginLimiter) recent(key string, now time.Time) []time.Time {
	times := l.fails[key]
	i := 0
	for i < len(times) && !times[i].After(now.Add(-loginWindow)) {
		i++
	}
	times = times[i:]
	if len(times) == 0 {
		delete(l.fails, key)
		return nil
	}
	l.fails[key] = times
	return times
}

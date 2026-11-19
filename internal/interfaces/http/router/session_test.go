package routes

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// Сессия хранится в базе: после перезапуска (новый роутер на той же базе) вход сохраняется,
// а после выхода кука больше не действует.
func TestSessionSurvivesRestartAndEndsOnLogout(t *testing.T) {
	c, srv := newTestApp(t)
	id := addUser(t, c, "ivanov@example.com", "lead_specialist", "epidemiological_supervision")
	client := login(t, srv, "ivanov@example.com")

	srvURL, _ := url.Parse(srv.URL)
	cookies := client.Jar.Cookies(srvURL)
	if len(cookies) == 0 {
		t.Fatal("no session cookie")
	}

	router, err := NewRouter(c, noopLogger{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := router.Build()
	if err != nil {
		t.Fatal(err)
	}
	restarted := httptest.NewServer(handler)
	defer restarted.Close()

	noRedirect := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	request := func(base string) int {
		req, _ := http.NewRequest(http.MethodGet, base+"/users/"+id, nil)
		for _, ck := range cookies {
			req.AddCookie(ck)
		}
		res, err := noRedirect.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		return res.StatusCode
	}
	if got := request(restarted.URL); got != http.StatusOK {
		t.Fatalf("after restart: status %d, want 200", got)
	}

	postForm(t, client, srv.URL+"/logout", url.Values{}, http.StatusOK)
	if got := request(restarted.URL); got != http.StatusSeeOther {
		t.Fatalf("after logout: status %d, want redirect to login", got)
	}
}

// После пяти неверных паролей вход блокируется даже с верным паролем.
func TestLoginBlockedAfterFailedAttempts(t *testing.T) {
	c, srv := newTestApp(t)
	addUser(t, c, "ivanov@example.com", "lead_specialist", "epidemiological_supervision")

	attempt := func(password string) int {
		res, err := http.PostForm(srv.URL+"/login", url.Values{"email": {"ivanov@example.com"}, "password": {password}})
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		return res.StatusCode
	}
	for i := 0; i < 5; i++ {
		if got := attempt("wrong-password"); got != http.StatusUnauthorized {
			t.Fatalf("attempt %d: status %d, want 401", i+1, got)
		}
	}
	if got := attempt(testPassword); got != http.StatusTooManyRequests {
		t.Fatalf("after 5 failures: status %d, want 429", got)
	}
}

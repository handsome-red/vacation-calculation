package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSameOrigin(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h := SameOrigin(ok)

	cases := []struct {
		name, method string
		headers      map[string]string
		want         int
	}{
		{"get from anywhere", http.MethodGet, map[string]string{"Origin": "https://evil.example"}, http.StatusOK},
		{"post without browser headers", http.MethodPost, nil, http.StatusOK},
		{"post from same origin", http.MethodPost, map[string]string{"Origin": "http://app.local", "Sec-Fetch-Site": "same-origin"}, http.StatusOK},
		{"post from other site", http.MethodPost, map[string]string{"Origin": "https://evil.example"}, http.StatusForbidden},
		{"delete cross-site", http.MethodDelete, map[string]string{"Sec-Fetch-Site": "cross-site"}, http.StatusForbidden},
	}
	for _, c := range cases {
		req := httptest.NewRequest(c.method, "http://app.local/users/1/edit", nil)
		for k, v := range c.headers {
			req.Header.Set(k, v)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != c.want {
			t.Errorf("%s: status %d, want %d", c.name, rec.Code, c.want)
		}
	}
}

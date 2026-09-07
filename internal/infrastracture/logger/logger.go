package logger

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/buffer"
)

func TestLogger(t *testing.T) {
	t.Run("logs request details", func(t *testing.T) {
		buf := &buffer.Buffer{}

		log := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))

		text := "test response"

		handler := Logger(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(text))
		}))

		req := httptest.NewRequest(http.MethodGet, "/test/path?param=value", nil)
		req.RemoteAddr = "192.168.1.100:12345"

		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, text, w.Body.String())

		logOutput := buf.String()
		assert.Contains(t, logOutput, "HTTP request completed")
		assert.Contains(t, logOutput, "method=GET")
		assert.Contains(t, logOutput, "path=/test/path")
		assert.Contains(t, logOutput, "status=200")
		assert.Contains(t, logOutput, "size=13")
		// assert.Contains(t, logOutput, "remote_addr=192.168.1.100")
		assert.Contains(t, logOutput, "request_id=")
	})

	// t.Run("logs error status", func(t *testing.T) {

	// })
}

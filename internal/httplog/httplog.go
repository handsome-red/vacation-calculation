package httplog

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"vacation-calculation/internal/types"
)

func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := generateRequestID()

			reqLog := log.With(
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				// slog.String("remote_addr", GetClientIP(r.Context())),
			)

			ctx := context.WithValue(r.Context(), types.LoggerContextKey, reqLog)

			rw := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
				size:           0,
			}

			next.ServeHTTP(rw, r.WithContext(ctx))

			reqLog.Info("HTTP request completed",
				slog.Int("status", rw.status),
				slog.Duration("duration", time.Since(start)),
				slog.Int("size", rw.size),
			)
		})
	}
}

func LoggerFromContext(ctx context.Context) *slog.Logger {

	if log, ok := ctx.Value(types.LoggerContextKey).(*slog.Logger); ok {
		return log
	}

	return slog.Default()
}

func GetClientIP(ctx context.Context) string {
	if ip, ok := ctx.Value(types.ClientIPContextKey).(string); ok {
		return ip
	}

	return "unknown"
}

func generateRequestID() string {
	now := time.Now().UnixNano()
	return fmt.Sprintf("%x-%d", now, now%10000)
}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

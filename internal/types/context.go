package types

type contextKey string

const (
	LoggerContextKey   contextKey = "logger"
	ClientIPContextKey contextKey = "client_ip"
)

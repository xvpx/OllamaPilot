package utils

import (
	"context"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	zerolog.Logger
}

// LoggerContextKey is the key used to store logger in context
type LoggerContextKey struct{}

// NewLogger creates a new logger instance
func NewLogger(level, format string) *Logger {
	// Parse log level
	logLevel, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	// Configure zerolog
	zerolog.SetGlobalLevel(logLevel)

	var logger zerolog.Logger

	// Configure output format
	switch strings.ToLower(format) {
	case "console", "text":
		logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	case "json":
		logger = zerolog.New(os.Stdout)
	default:
		logger = zerolog.New(os.Stdout)
	}

	// Add timestamp and caller info
	logger = logger.With().
		Timestamp().
		Caller().
		Logger()

	return &Logger{Logger: logger}
}

// WithContext adds logger to context
func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, LoggerContextKey{}, l)
}

// FromContext retrieves logger from context
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(LoggerContextKey{}).(*Logger); ok {
		return logger
	}
	// Return default logger if not found in context
	return &Logger{Logger: log.Logger}
}

// WithRequestID adds request ID to logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Logger: l.With().Str("request_id", requestID).Logger()}
}

// WithUserID adds user ID to logger
func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{Logger: l.With().Str("user_id", userID).Logger()}
}

// WithSessionID adds session ID to logger
func (l *Logger) WithSessionID(sessionID string) *Logger {
	return &Logger{Logger: l.With().Str("session_id", sessionID).Logger()}
}

// WithError adds error to logger
func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.With().Err(err).Logger()}
}

// WithComponent adds component name to logger
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{Logger: l.With().Str("component", component).Logger()}
}

// LogHTTPRequest logs HTTP request details
func (l *Logger) LogHTTPRequest(method, path, userAgent, remoteAddr string, statusCode int, duration int64) {
	l.Info().
		Str("method", method).
		Str("path", path).
		Str("user_agent", userAgent).
		Str("remote_addr", remoteAddr).
		Int("status_code", statusCode).
		Int64("duration_ms", duration).
		Msg("HTTP request")
}

// LogDatabaseQuery logs database query details
func (l *Logger) LogDatabaseQuery(query string, duration int64, err error) {
	event := l.Debug().
		Str("query", query).
		Int64("duration_ms", duration)

	if err != nil {
		event = event.Err(err)
	}

	event.Msg("Database query")
}

// LogOllamaRequest logs Ollama API request details
func (l *Logger) LogOllamaRequest(model, endpoint string, duration int64, err error) {
	event := l.Info().
		Str("model", model).
		Str("endpoint", endpoint).
		Int64("duration_ms", duration)

	if err != nil {
		event = event.Err(err)
	}

	event.Msg("Ollama request")
}

// LogStreamConnection logs streaming connection details
func (l *Logger) LogStreamConnection(sessionID, connectionID string, action string) {
	l.Info().
		Str("session_id", sessionID).
		Str("connection_id", connectionID).
		Str("action", action).
		Msg("Stream connection")
}

// LogPanic logs panic recovery details
func (l *Logger) LogPanic(recovered interface{}, stack []byte) {
	l.Error().
		Interface("panic", recovered).
		Bytes("stack", stack).
		Msg("Panic recovered")
}
package middleware

import (
	"net/http"
	"time"

	"chat_ollama/internal/utils"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(logger *utils.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID and add to context
			requestID := utils.GetRequestID(r)
			requestLogger := logger.WithRequestID(requestID)

			// Add logger to context
			ctx := requestLogger.WithContext(r.Context())
			r = r.WithContext(ctx)

			// Wrap response writer to capture status code
			rw := utils.NewResponseWriter(w)

			// Add request ID to response headers
			rw.Header().Set("X-Request-ID", requestID)

			// Process request
			next.ServeHTTP(rw, r)

			// Log request details
			duration := time.Since(start).Milliseconds()
			clientIP := utils.GetClientIP(r)
			userAgent := r.UserAgent()

			requestLogger.LogHTTPRequest(
				r.Method,
				r.URL.Path,
				userAgent,
				clientIP,
				rw.StatusCode(),
				duration,
			)
		})
	}
}
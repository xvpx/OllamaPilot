package middleware

import (
	"net/http"
	"runtime/debug"

	"chat_ollama/internal/utils"
)

// RecoveryMiddleware recovers from panics and logs them
func RecoveryMiddleware(logger *utils.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					// Get logger from context or use the provided one
					requestLogger := utils.FromContext(r.Context())
					if requestLogger == nil {
						requestLogger = logger
					}

					// Log the panic with stack trace
					stack := debug.Stack()
					requestLogger.LogPanic(recovered, stack)

					// Create error response
					apiErr := utils.NewInternalError(
						"An unexpected error occurred",
						r.URL.Path,
					)

					// Write error response
					utils.WriteError(w, apiErr)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
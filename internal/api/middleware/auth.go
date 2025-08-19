package middleware

import (
	"context"
	"net/http"
	"strings"

	"chat_ollama/internal/models"
	"chat_ollama/internal/services"
	"chat_ollama/internal/utils"
)

// AuthContextKey is the key used to store auth context in request context
type AuthContextKey string

const (
	// UserContextKey is the key for storing user context
	UserContextKey AuthContextKey = "user"
)

// AuthMiddleware creates middleware for JWT authentication
func AuthMiddleware(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				apiErr := utils.NewUnauthorizedError("Authorization header required", "authorization")
				utils.WriteError(w, apiErr)
				return
			}

			// Check if header starts with "Bearer "
			if !strings.HasPrefix(authHeader, "Bearer ") {
				apiErr := utils.NewUnauthorizedError("Invalid authorization header format", "authorization")
				utils.WriteError(w, apiErr)
				return
			}

			// Extract token
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				apiErr := utils.NewUnauthorizedError("Token required", "token")
				utils.WriteError(w, apiErr)
				return
			}

			// Validate token
			authContext, err := authService.ValidateJWT(token)
			if err != nil {
				apiErr := utils.NewUnauthorizedError("Invalid or expired token", "token")
				utils.WriteError(w, apiErr)
				return
			}

			// Add auth context to request context
			ctx := context.WithValue(r.Context(), UserContextKey, authContext)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware creates middleware for optional JWT authentication
// This allows endpoints to work with or without authentication
func OptionalAuthMiddleware(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			
			// If no auth header, continue without authentication
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check if header starts with "Bearer "
			if !strings.HasPrefix(authHeader, "Bearer ") {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Validate token
			authContext, err := authService.ValidateJWT(token)
			if err != nil {
				// If token is invalid, continue without authentication
				next.ServeHTTP(w, r)
				return
			}

			// Add auth context to request context
			ctx := context.WithValue(r.Context(), UserContextKey, authContext)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext extracts the authenticated user from request context
func GetUserFromContext(r *http.Request) (*models.AuthContext, bool) {
	authContext, ok := r.Context().Value(UserContextKey).(*models.AuthContext)
	return authContext, ok
}

// RequireAuth is a helper function to check if user is authenticated
func RequireAuth(r *http.Request) (*models.AuthContext, error) {
	authContext, ok := GetUserFromContext(r)
	if !ok {
		return nil, utils.NewUnauthorizedError("Authentication required", "authentication")
	}
	return authContext, nil
}
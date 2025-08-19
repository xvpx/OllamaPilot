package utils

import (
	"fmt"
	"time"
)

// ErrorType represents different types of errors
type ErrorType string

const (
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeInternal   ErrorType = "internal"
	ErrorTypeOllama     ErrorType = "ollama"
	ErrorTypeDatabase   ErrorType = "database"
	ErrorTypeRateLimit  ErrorType = "rate_limit"
)

// APIError represents a structured API error following RFC 7807
type APIError struct {
	Type      ErrorType `json:"type"`
	Title     string    `json:"title"`
	Status    int       `json:"status"`
	Detail    string    `json:"detail"`
	Instance  string    `json:"instance"`
	Timestamp time.Time `json:"timestamp"`
}

// Error implements the error interface
func (e APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Title, e.Detail)
}

// NewValidationError creates a new validation error
func NewValidationError(detail, instance string) APIError {
	return APIError{
		Type:      ErrorTypeValidation,
		Title:     "Validation Error",
		Status:    400,
		Detail:    detail,
		Instance:  instance,
		Timestamp: time.Now(),
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(detail, instance string) APIError {
	return APIError{
		Type:      ErrorTypeNotFound,
		Title:     "Not Found",
		Status:    404,
		Detail:    detail,
		Instance:  instance,
		Timestamp: time.Now(),
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(detail, instance string) APIError {
	return APIError{
		Type:      ErrorTypeInternal,
		Title:     "Internal Server Error",
		Status:    500,
		Detail:    detail,
		Instance:  instance,
		Timestamp: time.Now(),
	}
}

// NewOllamaError creates a new Ollama service error
func NewOllamaError(detail, instance string) APIError {
	return APIError{
		Type:      ErrorTypeOllama,
		Title:     "Ollama Service Error",
		Status:    502,
		Detail:    detail,
		Instance:  instance,
		Timestamp: time.Now(),
	}
}

// NewDatabaseError creates a new database error
func NewDatabaseError(detail, instance string) APIError {
	return APIError{
		Type:      ErrorTypeDatabase,
		Title:     "Database Error",
		Status:    500,
		Detail:    detail,
		Instance:  instance,
		Timestamp: time.Now(),
	}
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(detail, instance string) APIError {
	return APIError{
		Type:      ErrorTypeRateLimit,
		Title:     "Rate Limit Exceeded",
		Status:    429,
		Detail:    detail,
		Instance:  instance,
		Timestamp: time.Now(),
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(detail, instance string) APIError {
	return APIError{
		Type:      ErrorTypeValidation, // Using validation type since there's no unauthorized type
		Title:     "Unauthorized",
		Status:    401,
		Detail:    detail,
		Instance:  instance,
		Timestamp: time.Now(),
	}
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(detail, instance string) APIError {
	return APIError{
		Type:      ErrorTypeValidation, // Using validation type since there's no forbidden type
		Title:     "Forbidden",
		Status:    403,
		Detail:    detail,
		Instance:  instance,
		Timestamp: time.Now(),
	}
}

// WrapError wraps a generic error into an APIError
func WrapError(err error, errorType ErrorType, instance string) APIError {
	var title string
	var status int

	switch errorType {
	case ErrorTypeValidation:
		title = "Validation Error"
		status = 400
	case ErrorTypeNotFound:
		title = "Not Found"
		status = 404
	case ErrorTypeOllama:
		title = "Ollama Service Error"
		status = 502
	case ErrorTypeDatabase:
		title = "Database Error"
		status = 500
	case ErrorTypeRateLimit:
		title = "Rate Limit Exceeded"
		status = 429
	default:
		title = "Internal Server Error"
		status = 500
		errorType = ErrorTypeInternal
	}

	return APIError{
		Type:      errorType,
		Title:     title,
		Status:    status,
		Detail:    err.Error(),
		Instance:  instance,
		Timestamp: time.Now(),
	}
}
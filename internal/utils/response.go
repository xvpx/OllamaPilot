package utils

import (
	"encoding/json"
	"net/http"
	"time"
)

// ResponseWriter wraps http.ResponseWriter with additional functionality
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriter creates a new ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// StatusCode returns the captured status code
func (rw *ResponseWriter) StatusCode() int {
	return rw.statusCode
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, err APIError) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	return json.NewEncoder(w).Encode(err)
}

// WriteSuccess writes a success response
func WriteSuccess(w http.ResponseWriter, data interface{}) error {
	return WriteJSON(w, http.StatusOK, data)
}

// WriteCreated writes a created response
func WriteCreated(w http.ResponseWriter, data interface{}) error {
	return WriteJSON(w, http.StatusCreated, data)
}

// WriteNoContent writes a no content response
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// WriteSSEHeaders sets headers for Server-Sent Events
func WriteSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")
}

// WriteSSEData writes SSE data
func WriteSSEData(w http.ResponseWriter, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte("data: " + string(jsonData) + "\n\n"))
	if err != nil {
		return err
	}

	// Flush the data immediately
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// WriteSSEEvent writes SSE event with custom event type
func WriteSSEEvent(w http.ResponseWriter, eventType string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte("event: " + eventType + "\n"))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte("data: " + string(jsonData) + "\n\n"))
	if err != nil {
		return err
	}

	// Flush the data immediately
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// WriteSSEComment writes SSE comment (for keep-alive)
func WriteSSEComment(w http.ResponseWriter, comment string) error {
	_, err := w.Write([]byte(": " + comment + "\n\n"))
	if err != nil {
		return err
	}

	// Flush the data immediately
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// ParseJSON parses JSON from request body
func ParseJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// GetRequestID extracts request ID from headers or generates one
func GetRequestID(r *http.Request) string {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = r.Header.Get("X-Correlation-ID")
	}
	if requestID == "" {
		// Generate a simple request ID if none provided
		// In production, you might want to use a proper UUID library
		requestID = "req-" + generateSimpleID()
	}
	return requestID
}

// generateSimpleID generates a simple ID (placeholder implementation)
func generateSimpleID() string {
	// This is a simple implementation - in production use proper UUID
	return "12345678"
}

// GetClientIP extracts client IP from request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// GetCurrentTime returns the current time
func GetCurrentTime() time.Time {
	return time.Now()
}
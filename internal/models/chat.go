package models

import (
	"time"
)

// Message represents a chat message
type Message struct {
	ID         string    `json:"id" db:"id"`
	SessionID  string    `json:"session_id" db:"session_id"`
	Role       string    `json:"role" db:"role"`
	Content    string    `json:"content" db:"content"`
	Model      string    `json:"model,omitempty" db:"model"`
	TokensUsed int       `json:"tokens_used,omitempty" db:"tokens_used"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Message   string                 `json:"message"`
	SessionID string                 `json:"session_id"`
	Model     string                 `json:"model"`
	Stream    bool                   `json:"stream"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// ChatResponse represents a non-streaming chat response
type ChatResponse struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"session_id"`
	Content    string    `json:"content"`
	Model      string    `json:"model"`
	CreatedAt  time.Time `json:"created_at"`
	TokensUsed int       `json:"tokens_used"`
}

// StreamResponse represents a streaming chat response
type StreamResponse struct {
	Type      string                 `json:"type"`
	Content   string                 `json:"content,omitempty"`
	SessionID string                 `json:"session_id"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MessagesResponse represents the response for getting conversation history
type MessagesResponse struct {
	SessionID string    `json:"session_id"`
	Messages  []Message `json:"messages"`
}

// ValidateRole validates if the role is valid
func ValidateRole(role string) bool {
	switch role {
	case "user", "assistant", "system":
		return true
	default:
		return false
	}
}
package models

import (
	"time"
)

// Session represents a chat session
type Session struct {
	ID           string    `json:"id" db:"id"`
	Title        string    `json:"title" db:"title"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	MessageCount int       `json:"message_count,omitempty"`
}

// SessionsResponse represents the response for listing sessions
type SessionsResponse struct {
	Sessions []Session `json:"sessions"`
}
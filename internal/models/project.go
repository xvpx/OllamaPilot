package models

import (
	"time"
)

// Project represents a project that can contain multiple chat sessions
type Project struct {
	ID          string    `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description,omitempty" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ProjectsResponse represents the response for listing projects
type ProjectsResponse struct {
	Projects []Project `json:"projects"`
}

// CreateProjectRequest represents a request to create a new project
type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty" validate:"max=1000"`
}

// UpdateProjectRequest represents a request to update a project
type UpdateProjectRequest struct {
	Name        string `json:"name,omitempty" validate:"min=1,max=255"`
	Description string `json:"description,omitempty" validate:"max=1000"`
	IsActive    *bool  `json:"is_active,omitempty"`
}
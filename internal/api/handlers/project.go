package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"chat_ollama/internal/config"
	"chat_ollama/internal/database"
	"chat_ollama/internal/models"
	"chat_ollama/internal/utils"
)

// ProjectHandler handles project-related requests
type ProjectHandler struct {
	db     database.Database
	logger *utils.Logger
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(db database.Database, cfg *config.Config, logger *utils.Logger) *ProjectHandler {
	return &ProjectHandler{
		db:     db,
		logger: logger.WithComponent("project_handler"),
	}
}

// GetProjects handles GET /v1/projects
func (h *ProjectHandler) GetProjects(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Msg("Getting projects list")

	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
	`

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to query projects")
		apiErr := utils.NewInternalError("Failed to retrieve projects", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.IsActive,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to scan project row")
			apiErr := utils.NewInternalError("Failed to process projects", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating project rows")
		apiErr := utils.NewInternalError("Failed to retrieve projects", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	response := models.ProjectsResponse{
		Projects: projects,
	}

	logger.Info().Int("project_count", len(projects)).Msg("Projects retrieved successfully")
	utils.WriteSuccess(w, response)
}

// GetProject handles GET /v1/projects/{projectID}
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		apiErr := utils.NewValidationError("Project ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Str("project_id", projectID).Msg("Getting project")

	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	var project models.Project
	err := h.db.QueryRowContext(ctx, query, projectID).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.IsActive,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			apiErr := utils.NewNotFoundError("Project not found", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		logger.Error().Err(err).Str("project_id", projectID).Msg("Failed to get project")
		apiErr := utils.NewInternalError("Failed to retrieve project", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().Str("project_id", projectID).Msg("Project retrieved successfully")
	utils.WriteSuccess(w, project)
}

// CreateProject handles POST /v1/projects
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	var req models.CreateProjectRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		logger.Error().Err(err).Msg("Failed to parse create project request")
		apiErr := utils.NewValidationError("Invalid JSON in request body", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	// Validate request
	if req.Name == "" {
		apiErr := utils.NewValidationError("Name field is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Generate project ID
	projectID := uuid.New().String()

	logger.Info().
		Str("project_id", projectID).
		Str("name", req.Name).
		Msg("Creating project")

	query := `
		INSERT INTO projects (id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, description, is_active, created_at, updated_at
	`

	var project models.Project
	err := h.db.QueryRowContext(ctx, query, projectID, req.Name, req.Description, true).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.IsActive,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		logger.Error().Err(err).Str("project_id", projectID).Msg("Failed to create project")
		apiErr := utils.NewInternalError("Failed to create project", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().
		Str("project_id", projectID).
		Str("name", req.Name).
		Msg("Project created successfully")

	utils.WriteSuccess(w, project)
}

// UpdateProject handles PUT /v1/projects/{projectID}
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		apiErr := utils.NewValidationError("Project ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	var req models.UpdateProjectRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		logger.Error().Err(err).Msg("Failed to parse update project request")
		apiErr := utils.NewValidationError("Invalid JSON in request body", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().
		Str("project_id", projectID).
		Msg("Updating project")

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != "" {
		setParts = append(setParts, "name = $"+strconv.Itoa(argIndex))
		args = append(args, req.Name)
		argIndex++
	}

	if req.Description != "" {
		setParts = append(setParts, "description = $"+strconv.Itoa(argIndex))
		args = append(args, req.Description)
		argIndex++
	}

	if req.IsActive != nil {
		setParts = append(setParts, "is_active = $"+strconv.Itoa(argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(setParts) == 0 {
		apiErr := utils.NewValidationError("No fields to update", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, projectID)

	query := fmt.Sprintf(`
		UPDATE projects
		SET %s
		WHERE id = $%d
		RETURNING id, name, description, is_active, created_at, updated_at
	`, strings.Join(setParts, ", "), argIndex)

	var project models.Project
	err := h.db.QueryRowContext(ctx, query, args...).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.IsActive,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			apiErr := utils.NewNotFoundError("Project not found", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		logger.Error().Err(err).Str("project_id", projectID).Msg("Failed to update project")
		apiErr := utils.NewInternalError("Failed to update project", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().
		Str("project_id", projectID).
		Msg("Project updated successfully")

	utils.WriteSuccess(w, project)
}

// DeleteProject handles DELETE /v1/projects/{projectID}
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		apiErr := utils.NewValidationError("Project ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().
		Str("project_id", projectID).
		Msg("Deleting project")

	// First check if project exists
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)"
	err := h.db.QueryRowContext(ctx, checkQuery, projectID).Scan(&exists)
	if err != nil {
		logger.Error().Err(err).Str("project_id", projectID).Msg("Failed to check project existence")
		apiErr := utils.NewInternalError("Failed to delete project", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	if !exists {
		apiErr := utils.NewNotFoundError("Project not found", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	// Delete the project (this will set project_id to NULL for associated sessions due to ON DELETE SET NULL)
	deleteQuery := "DELETE FROM projects WHERE id = $1"
	result, err := h.db.ExecContext(ctx, deleteQuery, projectID)
	if err != nil {
		logger.Error().Err(err).Str("project_id", projectID).Msg("Failed to delete project")
		apiErr := utils.NewInternalError("Failed to delete project", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error().Err(err).Str("project_id", projectID).Msg("Failed to get rows affected")
		apiErr := utils.NewInternalError("Failed to delete project", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	if rowsAffected == 0 {
		apiErr := utils.NewNotFoundError("Project not found", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().
		Str("project_id", projectID).
		Msg("Project deleted successfully")

	// Return success response
	utils.WriteSuccess(w, map[string]string{
		"message":    "Project deleted successfully",
		"project_id": projectID,
	})
}

// GetProjectSessions handles GET /v1/projects/{projectID}/sessions
func (h *ProjectHandler) GetProjectSessions(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		apiErr := utils.NewValidationError("Project ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Str("project_id", projectID).Msg("Getting project sessions")

	// First check if project exists
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)"
	err := h.db.QueryRowContext(ctx, checkQuery, projectID).Scan(&exists)
	if err != nil {
		logger.Error().Err(err).Str("project_id", projectID).Msg("Failed to check project existence")
		apiErr := utils.NewInternalError("Failed to retrieve project sessions", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	if !exists {
		apiErr := utils.NewNotFoundError("Project not found", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	// Get sessions for this project
	query := `
		SELECT id, title, created_at, updated_at,
		       (SELECT COUNT(*) FROM messages WHERE session_id = sessions.id) as message_count
		FROM sessions
		WHERE project_id = $1
		ORDER BY updated_at DESC
	`

	rows, err := h.db.QueryContext(ctx, query, projectID)
	if err != nil {
		logger.Error().Err(err).Str("project_id", projectID).Msg("Failed to query project sessions")
		apiErr := utils.NewInternalError("Failed to retrieve project sessions", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		err := rows.Scan(
			&session.ID,
			&session.Title,
			&session.CreatedAt,
			&session.UpdatedAt,
			&session.MessageCount,
		)
		if err != nil {
			logger.Error().Err(err).Str("project_id", projectID).Msg("Failed to scan session row")
			apiErr := utils.NewInternalError("Failed to process project sessions", r.URL.Path)
			utils.WriteError(w, apiErr)
			return
		}
		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Str("project_id", projectID).Msg("Error iterating session rows")
		apiErr := utils.NewInternalError("Failed to retrieve project sessions", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	response := models.SessionsResponse{
		Sessions: sessions,
	}

	logger.Info().
		Str("project_id", projectID).
		Int("session_count", len(sessions)).
		Msg("Project sessions retrieved successfully")

	utils.WriteSuccess(w, response)
}
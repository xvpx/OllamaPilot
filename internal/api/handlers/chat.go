package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"chat_ollama/internal/api/middleware"
	"chat_ollama/internal/config"
	"chat_ollama/internal/database"
	"chat_ollama/internal/models"
	"chat_ollama/internal/services"
	"chat_ollama/internal/utils"
)

// ChatHandler handles chat-related requests
type ChatHandler struct {
	chatService    *services.ChatService
	semanticMemory *services.SemanticMemoryService
	logger         *utils.Logger
}

// NewChatHandler creates a new chat handler
func NewChatHandler(db database.Database, cfg *config.Config, logger *utils.Logger) *ChatHandler {
	// Create Ollama client
	ollamaClient := services.NewOllamaClient(cfg.OllamaHost, cfg.OllamaTimeout, logger)
	
	// Create embedding service
	embeddingService := services.NewEmbeddingService(cfg, logger)
	
	// Create chat service
	chatService := services.NewChatService(db, ollamaClient, embeddingService, cfg, logger)
	
	// Create semantic memory service with configured embedding model
	semanticMemory := services.NewSemanticMemoryServiceWithModel(db, embeddingService, logger, cfg.EmbeddingModel)

	return &ChatHandler{
		chatService:    chatService,
		semanticMemory: semanticMemory,
		logger:         logger.WithComponent("chat_handler"),
	}
}

// Chat handles POST /v1/chat
func (h *ChatHandler) Chat(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Get authenticated user from context (optional for debugging)
	authContext, ok := middleware.GetUserFromContext(r)
	if !ok {
		// For debugging: create a temporary auth context
		authContext = &models.AuthContext{
			UserID:   "debug-user-id",
			Username: "debug-user",
		}
		logger.Warn().Msg("No authentication context found, using debug user")
	}

	var req models.ChatRequest
	if err := utils.ParseJSON(r, &req); err != nil {
		logger.Error().Err(err).Msg("Failed to parse chat request")
		apiErr := utils.NewValidationError("Invalid JSON in request body", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	// Validate request
	if req.Message == "" {
		apiErr := utils.NewValidationError("Message field is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	if req.SessionID == "" {
		apiErr := utils.NewValidationError("Session ID field is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	// Verify session belongs to user or create it if it doesn't exist (skip for debug user)
	if authContext.UserID != "debug-user-id" {
		if err := h.verifyOrCreateSession(req.SessionID, authContext.UserID); err.Type != "" {
			utils.WriteError(w, err)
			return
		}
	} else {
		logger.Info().Str("session_id", req.SessionID).Msg("Skipping session ownership verification for debug user")
	}

	// Require model parameter
	if req.Model == "" {
		apiErr := utils.NewValidationError("Model field is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().
		Str("session_id", req.SessionID).
		Str("user_id", authContext.UserID).
		Str("model", req.Model).
		Bool("stream", req.Stream).
		Msg("Chat request received")

	if req.Stream {
		h.handleStreamingChat(w, r, req, authContext.UserID)
	} else {
		h.handleNonStreamingChat(w, r, req, authContext.UserID)
	}
}

// handleStreamingChat handles streaming chat requests
func (h *ChatHandler) handleStreamingChat(w http.ResponseWriter, r *http.Request, req models.ChatRequest, userID string) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Set SSE headers
	utils.WriteSSEHeaders(w)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	logger.Info().
		Str("session_id", req.SessionID).
		Msg("Starting streaming chat response")

	// Create response channel
	responseChan := make(chan models.StreamResponse, 100)

	// Start streaming chat processing
	go func() {
		if err := h.chatService.ProcessStreamingChatWithUser(ctx, req, userID, responseChan); err != nil {
			logger.Error().Err(err).
				Str("session_id", req.SessionID).
				Msg("Streaming chat processing failed")
		}
	}()

	// Stream responses to client
	for {
		select {
		case response, ok := <-responseChan:
			if !ok {
				// Channel closed, streaming complete
				return
			}

			if err := utils.WriteSSEData(w, response); err != nil {
				logger.Error().Err(err).Msg("Failed to write SSE data")
				return
			}

			// If error or done, stop streaming
			if response.Type == "error" || response.Type == "done" {
				logger.Info().
					Str("session_id", req.SessionID).
					Str("type", response.Type).
					Msg("Streaming chat response completed")
				return
			}

		case <-ctx.Done():
			logger.Warn().
				Str("session_id", req.SessionID).
				Msg("Streaming chat request timed out")
			
			errorResponse := models.StreamResponse{
				Type:      "error",
				SessionID: req.SessionID,
				Error:     "Request timed out",
			}
			utils.WriteSSEData(w, errorResponse)
			return
		}
	}
}

// handleNonStreamingChat handles non-streaming chat requests
func (h *ChatHandler) handleNonStreamingChat(w http.ResponseWriter, r *http.Request, req models.ChatRequest, userID string) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	logger.Info().
		Str("session_id", req.SessionID).
		Msg("Processing non-streaming chat request")

	response, err := h.chatService.ProcessChatWithUser(ctx, req, userID)
	if err != nil {
		logger.Error().Err(err).
			Str("session_id", req.SessionID).
			Msg("Chat processing failed")

		apiErr := utils.NewInternalError("Failed to process chat request", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().
		Str("session_id", req.SessionID).
		Str("message_id", response.ID).
		Int("tokens_used", response.TokensUsed).
		Msg("Non-streaming chat response completed")

	utils.WriteSuccess(w, response)
}

// GetSessions handles GET /v1/sessions
func (h *ChatHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Get authenticated user from context (optional for debugging)
	authContext, ok := middleware.GetUserFromContext(r)
	if !ok {
		// For debugging: create a temporary auth context
		authContext = &models.AuthContext{
			UserID:   "debug-user-id",
			Username: "debug-user",
		}
		logger.Warn().Msg("No authentication context found for sessions, using debug user")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Str("user_id", authContext.UserID).Msg("Getting sessions list")

	sessions, err := h.chatService.GetSessionsByUser(ctx, authContext.UserID)
	if err != nil {
		logger.Error().Err(err).Str("user_id", authContext.UserID).Msg("Failed to get sessions")
		apiErr := utils.NewInternalError("Failed to retrieve sessions", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	response := models.SessionsResponse{
		Sessions: sessions,
	}

	logger.Info().Str("user_id", authContext.UserID).Int("session_count", len(sessions)).Msg("Sessions retrieved successfully")
	utils.WriteSuccess(w, response)
}

// GetSessionMessages handles GET /v1/sessions/{id}/messages
func (h *ChatHandler) GetSessionMessages(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Get authenticated user from context (optional for debugging)
	authContext, ok := middleware.GetUserFromContext(r)
	if !ok {
		// For debugging: create a temporary auth context
		authContext = &models.AuthContext{
			UserID:   "debug-user-id",
			Username: "debug-user",
		}
		logger.Warn().Msg("No authentication context found for session messages, using debug user")
	}

	// Extract session ID from URL path
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		apiErr := utils.NewValidationError("Session ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	// Verify session belongs to user (skip for debug user)
	if authContext.UserID != "debug-user-id" {
		if err := h.verifySessionOwnership(sessionID, authContext.UserID); err.Type != "" {
			utils.WriteError(w, err)
			return
		}
	} else {
		logger.Info().Str("session_id", sessionID).Msg("Skipping session ownership verification for debug user")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().
		Str("session_id", sessionID).
		Str("user_id", authContext.UserID).
		Msg("Getting session messages")

	messages, err := h.chatService.GetSessionMessages(ctx, sessionID)
	if err != nil {
		logger.Error().Err(err).
			Str("session_id", sessionID).
			Str("user_id", authContext.UserID).
			Msg("Failed to get session messages")
		apiErr := utils.NewInternalError("Failed to retrieve session messages", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	response := models.MessagesResponse{
		SessionID: sessionID,
		Messages:  messages,
	}

	logger.Info().
		Str("session_id", sessionID).
		Str("user_id", authContext.UserID).
		Int("message_count", len(messages)).
		Msg("Session messages retrieved successfully")

	utils.WriteSuccess(w, response)
}

// DeleteSession handles DELETE /v1/sessions/{sessionID}
func (h *ChatHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Get authenticated user from context (optional for debugging)
	authContext, ok := middleware.GetUserFromContext(r)
	if !ok {
		// For debugging: create a temporary auth context
		authContext = &models.AuthContext{
			UserID:   "debug-user-id",
			Username: "debug-user",
		}
		logger.Warn().Msg("No authentication context found for session deletion, using debug user")
	}

	// Extract session ID from URL path
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		apiErr := utils.NewValidationError("Session ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	// Verify session belongs to user or handle legacy sessions
	if authContext.UserID != "debug-user-id" {
		if err := h.verifyOrClaimSession(sessionID, authContext.UserID); err.Type != "" {
			utils.WriteError(w, err)
			return
		}
	} else {
		logger.Info().Str("session_id", sessionID).Msg("Skipping session ownership verification for debug user")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().
		Str("session_id", sessionID).
		Str("user_id", authContext.UserID).
		Msg("Deleting session")

	err := h.chatService.DeleteSession(ctx, sessionID)
	if err != nil {
		logger.Error().Err(err).
			Str("session_id", sessionID).
			Str("user_id", authContext.UserID).
			Msg("Failed to delete session")
		apiErr := utils.NewInternalError("Failed to delete session", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().
		Str("session_id", sessionID).
		Str("user_id", authContext.UserID).
		Msg("Session deleted successfully")

	// Return success response
	utils.WriteSuccess(w, map[string]string{
		"message": "Session deleted successfully",
		"session_id": sessionID,
	})
}

// SearchMemory handles POST /v1/memory/search
func (h *ChatHandler) SearchMemory(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	var req struct {
		Query     string `json:"query"`
		SessionID string `json:"session_id,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}

	if err := utils.ParseJSON(r, &req); err != nil {
		logger.Error().Err(err).Msg("Failed to parse memory search request")
		apiErr := utils.NewValidationError("Invalid JSON in request body", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	if req.Query == "" {
		apiErr := utils.NewValidationError("Query field is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	if req.Limit <= 0 {
		req.Limit = 10 // Default limit
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().
		Str("query", req.Query).
		Str("session_id", req.SessionID).
		Int("limit", req.Limit).
		Msg("Searching semantic memory")

	results, err := h.semanticMemory.SearchSimilarMessages(ctx, req.Query, req.Limit, req.SessionID)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to search semantic memory")
		apiErr := utils.NewInternalError("Failed to search memory", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().
		Str("query", req.Query).
		Int("results_count", len(results)).
		Msg("Memory search completed")

	utils.WriteSuccess(w, map[string]interface{}{
		"query":   req.Query,
		"results": results,
		"count":   len(results),
	})
}

// GetMemorySummaries handles GET /v1/memory/summaries
func (h *ChatHandler) GetMemorySummaries(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	sessionID := r.URL.Query().Get("session_id")
	summaryType := r.URL.Query().Get("type")

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().
		Str("session_id", sessionID).
		Str("summary_type", summaryType).
		Msg("Getting memory summaries")

	summaries, err := h.semanticMemory.GetMemorySummaries(ctx, sessionID, summaryType)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get memory summaries")
		apiErr := utils.NewInternalError("Failed to retrieve memory summaries", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().
		Int("summaries_count", len(summaries)).
		Msg("Memory summaries retrieved")

	utils.WriteSuccess(w, map[string]interface{}{
		"summaries": summaries,
		"count":     len(summaries),
	})
}

// CreateMemorySummary handles POST /v1/memory/summaries
func (h *ChatHandler) CreateMemorySummary(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	var req struct {
		SessionID    string `json:"session_id"`
		SummaryType  string `json:"summary_type"`
		Title        string `json:"title,omitempty"`
		Content      string `json:"content"`
		MessageCount int    `json:"message_count,omitempty"`
	}

	if err := utils.ParseJSON(r, &req); err != nil {
		logger.Error().Err(err).Msg("Failed to parse memory summary request")
		apiErr := utils.NewValidationError("Invalid JSON in request body", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	if req.Content == "" {
		apiErr := utils.NewValidationError("Content field is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	if req.SummaryType == "" {
		req.SummaryType = "conversation"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().
		Str("session_id", req.SessionID).
		Str("summary_type", req.SummaryType).
		Msg("Creating memory summary")

	summary, err := h.semanticMemory.CreateMemorySummary(
		ctx,
		req.SessionID,
		req.SummaryType,
		req.Title,
		req.Content,
		time.Now(),
		time.Now(),
		req.MessageCount,
	)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create memory summary")
		apiErr := utils.NewInternalError("Failed to create memory summary", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().
		Str("summary_id", summary.ID).
		Str("session_id", req.SessionID).
		Msg("Memory summary created")

	utils.WriteSuccess(w, summary)
}

// GetMemoryGaps handles GET /v1/memory/gaps/{sessionID}
func (h *ChatHandler) GetMemoryGaps(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		apiErr := utils.NewValidationError("Session ID is required", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().
		Str("session_id", sessionID).
		Msg("Detecting memory gaps")

	// Default threshold of 1 hour
	threshold := 1 * time.Hour
	if thresholdParam := r.URL.Query().Get("threshold"); thresholdParam != "" {
		if parsedThreshold, err := time.ParseDuration(thresholdParam); err == nil {
			threshold = parsedThreshold
		}
	}

	gaps, err := h.semanticMemory.DetectMemoryGaps(ctx, sessionID, threshold)
	if err != nil {
		logger.Error().Err(err).
			Str("session_id", sessionID).
			Msg("Failed to detect memory gaps")
		apiErr := utils.NewInternalError("Failed to detect memory gaps", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().
		Str("session_id", sessionID).
		Int("gaps_count", len(gaps)).
		Msg("Memory gaps detected")

	utils.WriteSuccess(w, map[string]interface{}{
		"session_id": sessionID,
		"gaps":       gaps,
		"count":      len(gaps),
		"threshold":  threshold.String(),
	})
}

// verifySessionOwnership checks if a session belongs to the specified user
func (h *ChatHandler) verifySessionOwnership(sessionID, userID string) utils.APIError {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	session, err := h.chatService.GetSessionByID(ctx, sessionID)
	if err != nil {
		h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get session for ownership verification")
		return utils.NewNotFoundError("Session not found", sessionID)
	}

	if session.UserID != userID {
		h.logger.Warn().
			Str("session_id", sessionID).
			Str("user_id", userID).
			Str("session_user_id", session.UserID).
			Msg("User attempted to access session they don't own")
		return utils.NewForbiddenError("Access denied to this session", sessionID)
	}

	return utils.APIError{}
}

// verifyOrCreateSession checks if a session belongs to the specified user, or creates it if it doesn't exist
func (h *ChatHandler) verifyOrCreateSession(sessionID, userID string) utils.APIError {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	session, err := h.chatService.GetSessionByID(ctx, sessionID)
	if err != nil {
		// Session doesn't exist, create it
		h.logger.Info().
			Str("session_id", sessionID).
			Str("user_id", userID).
			Msg("Session not found, creating new session")
		
		newSession := models.Session{
			ID:        sessionID,
			UserID:    userID,
			Title:     "New Chat",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		createErr := h.chatService.CreateSession(ctx, newSession)
		if createErr != nil {
			h.logger.Error().Err(createErr).
				Str("session_id", sessionID).
				Str("user_id", userID).
				Msg("Failed to create new session")
			return utils.NewInternalError("Failed to create session", sessionID)
		}
		
		h.logger.Info().
			Str("session_id", sessionID).
			Str("user_id", userID).
			Msg("New session created successfully")
		return utils.APIError{}
	}

	// Session exists, verify ownership
	if session.UserID != userID {
		h.logger.Warn().
			Str("session_id", sessionID).
			Str("user_id", userID).
			Str("session_user_id", session.UserID).
			Msg("User attempted to access session they don't own")
		return utils.NewForbiddenError("Access denied to this session", sessionID)
	}

	return utils.APIError{}
}

// verifyOrClaimSession checks if a session belongs to the specified user, or claims it if it's a legacy debug session
func (h *ChatHandler) verifyOrClaimSession(sessionID, userID string) utils.APIError {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	session, err := h.chatService.GetSessionByID(ctx, sessionID)
	if err != nil {
		h.logger.Error().Err(err).Str("session_id", sessionID).Msg("Failed to get session for ownership verification")
		return utils.NewNotFoundError("Session not found", sessionID)
	}

	// If session belongs to current user, allow deletion
	if session.UserID == userID {
		return utils.APIError{}
	}

	// If session belongs to debug user (legacy session), allow current user to claim and delete it
	if session.UserID == "debug-user-id" {
		h.logger.Info().
			Str("session_id", sessionID).
			Str("user_id", userID).
			Str("previous_user_id", session.UserID).
			Msg("Claiming legacy debug session for deletion")
		
		// Update session ownership to current user before deletion
		// This ensures proper audit trail and allows deletion
		updateErr := h.chatService.UpdateSessionOwnership(ctx, sessionID, userID)
		if updateErr != nil {
			h.logger.Error().Err(updateErr).
				Str("session_id", sessionID).
				Str("user_id", userID).
				Msg("Failed to claim legacy session")
			return utils.NewInternalError("Failed to claim session for deletion", sessionID)
		}
		
		h.logger.Info().
			Str("session_id", sessionID).
			Str("user_id", userID).
			Msg("Successfully claimed legacy session")
		
		return utils.APIError{}
	}

	// Session belongs to a different real user, deny access
	h.logger.Warn().
		Str("session_id", sessionID).
		Str("user_id", userID).
		Str("session_user_id", session.UserID).
		Msg("User attempted to access session they don't own")
	return utils.NewForbiddenError("Access denied to this session", sessionID)
}
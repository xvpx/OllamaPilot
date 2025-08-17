package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

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

	// Set default model if not provided
	if req.Model == "" {
		req.Model = "llama2:7b" // Default model
	}

	logger.Info().
		Str("session_id", req.SessionID).
		Str("model", req.Model).
		Bool("stream", req.Stream).
		Msg("Chat request received")

	if req.Stream {
		h.handleStreamingChat(w, r, req)
	} else {
		h.handleNonStreamingChat(w, r, req)
	}
}

// handleStreamingChat handles streaming chat requests
func (h *ChatHandler) handleStreamingChat(w http.ResponseWriter, r *http.Request, req models.ChatRequest) {
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
		if err := h.chatService.ProcessStreamingChat(ctx, req, responseChan); err != nil {
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
func (h *ChatHandler) handleNonStreamingChat(w http.ResponseWriter, r *http.Request, req models.ChatRequest) {
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

	response, err := h.chatService.ProcessChat(ctx, req)
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

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	logger.Info().Msg("Getting sessions list")

	sessions, err := h.chatService.GetSessions(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get sessions")
		apiErr := utils.NewInternalError("Failed to retrieve sessions", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	response := models.SessionsResponse{
		Sessions: sessions,
	}

	logger.Info().Int("session_count", len(sessions)).Msg("Sessions retrieved successfully")
	utils.WriteSuccess(w, response)
}

// GetSessionMessages handles GET /v1/sessions/{id}/messages
func (h *ChatHandler) GetSessionMessages(w http.ResponseWriter, r *http.Request) {
	logger := utils.FromContext(r.Context())
	if logger == nil {
		logger = h.logger
	}

	// Extract session ID from URL path
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
		Msg("Getting session messages")

	messages, err := h.chatService.GetSessionMessages(ctx, sessionID)
	if err != nil {
		logger.Error().Err(err).
			Str("session_id", sessionID).
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

	// Extract session ID from URL path
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
		Msg("Deleting session")

	err := h.chatService.DeleteSession(ctx, sessionID)
	if err != nil {
		logger.Error().Err(err).
			Str("session_id", sessionID).
			Msg("Failed to delete session")
		apiErr := utils.NewInternalError("Failed to delete session", r.URL.Path)
		utils.WriteError(w, apiErr)
		return
	}

	logger.Info().
		Str("session_id", sessionID).
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

	logger.Info().
		Str("session_id", sessionID).
		Str("summary_type", summaryType).
		Msg("Getting memory summaries")

	// For now, return empty summaries as this requires database access
	// TODO: Implement proper memory summaries retrieval
	summaries := []interface{}{}

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
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
	chatService *services.ChatService
	logger      *utils.Logger
}

// NewChatHandler creates a new chat handler
func NewChatHandler(db *database.DB, cfg *config.Config, logger *utils.Logger) *ChatHandler {
	// Create Ollama client
	ollamaClient := services.NewOllamaClient(cfg.OllamaHost, cfg.OllamaTimeout, logger)
	
	// Create chat service
	chatService := services.NewChatService(db, ollamaClient, logger)

	return &ChatHandler{
		chatService: chatService,
		logger:      logger.WithComponent("chat_handler"),
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
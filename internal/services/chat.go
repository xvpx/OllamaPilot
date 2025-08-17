package services

import (
	"context"
	"fmt"
	"time"

	"chat_ollama/internal/config"
	"chat_ollama/internal/database"
	"chat_ollama/internal/models"
	"chat_ollama/internal/utils"

	"github.com/google/uuid"
)

// ChatService handles chat orchestration and message persistence
type ChatService struct {
	db             database.Database
	ollamaClient   *OllamaClient
	modelManager   *ModelManager
	semanticMemory *SemanticMemoryService
	logger         *utils.Logger
	config         *config.Config
}

// NewChatService creates a new chat service
func NewChatService(db database.Database, ollamaClient *OllamaClient, embeddingService *EmbeddingService, cfg *config.Config, logger *utils.Logger) *ChatService {
	// Create model manager - only works with SQLite for now
	var modelManager *ModelManager
	if sqliteDB, ok := db.(*database.DB); ok {
		modelManager = NewModelManager(sqliteDB, ollamaClient, logger)
	} else {
		// For PostgreSQL, we'll create a minimal model manager or skip it
		modelManager = nil
	}
	
	semanticMemory := NewSemanticMemoryServiceWithModel(db, embeddingService, logger, cfg.EmbeddingModel)
	
	return &ChatService{
		db:             db,
		ollamaClient:   ollamaClient,
		modelManager:   modelManager,
		semanticMemory: semanticMemory,
		logger:         logger.WithComponent("chat_service"),
		config:         cfg,
	}
}

// ProcessChat handles a non-streaming chat request
func (s *ChatService) ProcessChat(ctx context.Context, req models.ChatRequest) (*models.ChatResponse, error) {
	// Validate model if specified
	if req.Model != "" {
		if s.modelManager != nil {
			if err := s.modelManager.ValidateModel(ctx, req.Model); err != nil {
				return nil, fmt.Errorf("invalid model: %w", err)
			}
		}
	} else {
		// Use default model if none specified
		if s.modelManager != nil {
			defaultModel, err := s.modelManager.GetDefaultModel(ctx)
			if err != nil {
				s.logger.Warn().Err(err).Msg("No default model found, using fallback")
				req.Model = "llama2:7b" // Fallback model
			} else {
				req.Model = defaultModel.Name
			}
		} else {
			req.Model = "llama2:7b" // Fallback model when no model manager
		}
	}

	// Ensure session exists
	if err := s.ensureSession(ctx, req.SessionID); err != nil {
		return nil, fmt.Errorf("failed to ensure session: %w", err)
	}

	// Get conversation history
	messages, err := s.GetSessionMessages(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session messages: %w", err)
	}

	// Get relevant context from semantic memory if enabled
	var relevantContext string
	if s.config.EnableSemanticMemory && s.semanticMemory != nil {
		context, err := s.semanticMemory.GetRelevantContext(ctx, req.Message, req.SessionID, s.config.MaxContextResults)
		if err != nil {
			s.logger.Warn().Err(err).Msg("Failed to retrieve semantic context, continuing without it")
		} else if context != "" {
			relevantContext = context
			s.logger.Info().
				Str("session_id", req.SessionID).
				Int("context_length", len(context)).
				Int("max_results", s.config.MaxContextResults).
				Msg("Retrieved relevant context from semantic memory")
		}
	}

	s.logger.Info().
		Str("session_id", req.SessionID).
		Str("model", req.Model).
		Int("history_count", len(messages)).
		Bool("has_semantic_context", relevantContext != "").
		Msg("Processing chat request")

	// Save user message
	userMessage := models.Message{
		ID:        uuid.New().String(),
		SessionID: req.SessionID,
		Role:      "user",
		Content:   req.Message,
		CreatedAt: time.Now(),
	}

	if err := s.SaveMessage(ctx, userMessage); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Send to Ollama with semantic context
	ollamaResp, err := s.ollamaClient.ChatWithContext(ctx, req, messages, relevantContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get response from Ollama: %w", err)
	}

	// Save assistant message
	assistantMessage := models.Message{
		ID:         uuid.New().String(),
		SessionID:  req.SessionID,
		Role:       "assistant",
		Content:    ollamaResp.Message.Content,
		Model:      ollamaResp.Model,
		TokensUsed: ollamaResp.EvalCount,
		CreatedAt:  time.Now(),
	}

	if err := s.SaveMessage(ctx, assistantMessage); err != nil {
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}

	// Process messages for semantic memory (async)
	if s.semanticMemory != nil {
		go func() {
			if err := s.semanticMemory.ProcessMessageForSemanticMemory(context.Background(), userMessage); err != nil {
				s.logger.Error().Err(err).Str("message_id", userMessage.ID).Msg("Failed to process user message for semantic memory")
			}
			if err := s.semanticMemory.ProcessMessageForSemanticMemory(context.Background(), assistantMessage); err != nil {
				s.logger.Error().Err(err).Str("message_id", assistantMessage.ID).Msg("Failed to process assistant message for semantic memory")
			}
		}()
	}

	s.logger.Info().
		Str("session_id", req.SessionID).
		Str("message_id", assistantMessage.ID).
		Int("tokens_used", assistantMessage.TokensUsed).
		Msg("Chat request completed")

	return &models.ChatResponse{
		ID:         assistantMessage.ID,
		SessionID:  req.SessionID,
		Content:    assistantMessage.Content,
		Model:      assistantMessage.Model,
		CreatedAt:  assistantMessage.CreatedAt,
		TokensUsed: assistantMessage.TokensUsed,
	}, nil
}

// ProcessStreamingChat handles a streaming chat request
func (s *ChatService) ProcessStreamingChat(ctx context.Context, req models.ChatRequest, responseChan chan<- models.StreamResponse) error {
	defer close(responseChan)

	// Validate model if specified
	if req.Model != "" {
		if s.modelManager != nil {
			if err := s.modelManager.ValidateModel(ctx, req.Model); err != nil {
				responseChan <- models.StreamResponse{
					Type:      "error",
					SessionID: req.SessionID,
					Error:     fmt.Sprintf("Invalid model: %v", err),
				}
				return err
			}
		}
	} else {
		// Use default model if none specified
		if s.modelManager != nil {
			defaultModel, err := s.modelManager.GetDefaultModel(ctx)
			if err != nil {
				s.logger.Warn().Err(err).Msg("No default model found, using fallback")
				req.Model = "llama2:7b" // Fallback model
			} else {
				req.Model = defaultModel.Name
			}
		} else {
			req.Model = "llama2:7b" // Fallback model when no model manager
		}
	}

	// Ensure session exists
	if err := s.ensureSession(ctx, req.SessionID); err != nil {
		responseChan <- models.StreamResponse{
			Type:      "error",
			SessionID: req.SessionID,
			Error:     fmt.Sprintf("Failed to ensure session: %v", err),
		}
		return err
	}

	// Get conversation history
	messages, err := s.GetSessionMessages(ctx, req.SessionID)
	if err != nil {
		responseChan <- models.StreamResponse{
			Type:      "error",
			SessionID: req.SessionID,
			Error:     fmt.Sprintf("Failed to get session messages: %v", err),
		}
		return err
	}

	s.logger.Info().
		Str("session_id", req.SessionID).
		Str("model", req.Model).
		Int("history_count", len(messages)).
		Msg("Processing streaming chat request")

	// Save user message
	userMessage := models.Message{
		ID:        uuid.New().String(),
		SessionID: req.SessionID,
		Role:      "user",
		Content:   req.Message,
		CreatedAt: time.Now(),
	}

	if err := s.SaveMessage(ctx, userMessage); err != nil {
		responseChan <- models.StreamResponse{
			Type:      "error",
			SessionID: req.SessionID,
			Error:     fmt.Sprintf("Failed to save user message: %v", err),
		}
		return err
	}

	// Create channel for Ollama responses
	ollamaResponseChan := make(chan models.StreamResponse, 100)
	
	// Start streaming from Ollama with semantic context
	go func() {
		// Get relevant context for streaming as well if enabled
		var streamingContext string
		if s.config.EnableSemanticMemory && s.semanticMemory != nil {
			context, err := s.semanticMemory.GetRelevantContext(ctx, req.Message, req.SessionID, s.config.MaxContextResults)
			if err != nil {
				s.logger.Warn().Err(err).Msg("Failed to retrieve semantic context for streaming, continuing without it")
			} else {
				streamingContext = context
			}
		}
		
		if err := s.ollamaClient.ChatStreamWithContext(ctx, req, messages, streamingContext, ollamaResponseChan); err != nil {
			s.logger.Error().Err(err).Str("session_id", req.SessionID).Msg("Ollama streaming failed")
		}
	}()

	// Collect response content for saving
	var responseContent string
	var totalTokens int

	// Forward responses and collect content
	for ollamaResp := range ollamaResponseChan {
		if ollamaResp.Type == "token" {
			responseContent += ollamaResp.Content
		} else if ollamaResp.Type == "done" {
			if tokens, ok := ollamaResp.Metadata["total_tokens"].(int); ok {
				totalTokens = tokens
			}
		}

		// Forward to client
		responseChan <- ollamaResp

		// If error or done, break
		if ollamaResp.Type == "error" || ollamaResp.Type == "done" {
			break
		}
	}

	// Save assistant message if we got content
	if responseContent != "" {
		assistantMessage := models.Message{
			ID:         uuid.New().String(),
			SessionID:  req.SessionID,
			Role:       "assistant",
			Content:    responseContent,
			Model:      req.Model,
			TokensUsed: totalTokens,
			CreatedAt:  time.Now(),
		}

		if err := s.SaveMessage(ctx, assistantMessage); err != nil {
			s.logger.Error().Err(err).
				Str("session_id", req.SessionID).
				Msg("Failed to save assistant message")
		} else {
			s.logger.Info().
				Str("session_id", req.SessionID).
				Str("message_id", assistantMessage.ID).
				Int("tokens_used", totalTokens).
				Msg("Streaming chat request completed")
		}
	}

	return nil
}

// SaveMessage saves a message to the database
func (s *ChatService) SaveMessage(ctx context.Context, message models.Message) error {
	query := `
		INSERT INTO messages (id, session_id, role, content, model, tokens_used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.db.ExecContext(ctx, query,
		message.ID,
		message.SessionID,
		message.Role,
		message.Content,
		message.Model,
		message.TokensUsed,
		message.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	s.logger.Debug().
		Str("message_id", message.ID).
		Str("session_id", message.SessionID).
		Str("role", message.Role).
		Msg("Message saved to database")

	return nil
}

// GetSessionMessages retrieves all messages for a session
func (s *ChatService) GetSessionMessages(ctx context.Context, sessionID string) ([]models.Message, error) {
	query := `
		SELECT id, session_id, role, content, model, tokens_used, created_at
		FROM messages
		WHERE session_id = $1
		ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var model *string

		err := rows.Scan(
			&msg.ID,
			&msg.SessionID,
			&msg.Role,
			&msg.Content,
			&model,
			&msg.TokensUsed,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if model != nil {
			msg.Model = *model
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// ensureSession creates a session if it doesn't exist
func (s *ChatService) ensureSession(ctx context.Context, sessionID string) error {
	// Check if session exists
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM sessions WHERE id = $1)", sessionID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check session existence: %w", err)
	}

	if !exists {
		// Create new session
		session := models.Session{
			ID:        sessionID,
			Title:     "New Chat",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := s.CreateSession(ctx, session); err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		s.logger.Info().
			Str("session_id", sessionID).
			Msg("Created new session")
	}

	return nil
}

// CreateSession creates a new session
func (s *ChatService) CreateSession(ctx context.Context, session models.Session) error {
	query := `
		INSERT INTO sessions (id, title, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := s.db.ExecContext(ctx, query,
		session.ID,
		session.Title,
		session.CreatedAt,
		session.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSessions retrieves all sessions with message counts
func (s *ChatService) GetSessions(ctx context.Context) ([]models.Session, error) {
	query := `
		SELECT 
			s.id, 
			s.title, 
			s.created_at, 
			s.updated_at,
			COUNT(m.id) as message_count
		FROM sessions s
		LEFT JOIN messages m ON s.id = m.session_id
		GROUP BY s.id, s.title, s.created_at, s.updated_at
		ORDER BY s.updated_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
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
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// UpdateSessionTitle updates the title of a session
func (s *ChatService) UpdateSessionTitle(ctx context.Context, sessionID, title string) error {
	query := `
		UPDATE sessions
		SET title = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := s.db.ExecContext(ctx, query, title, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session title: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	s.logger.Info().
		Str("session_id", sessionID).
		Str("title", title).
		Msg("Session title updated")

	return nil
}

// DeleteSession deletes a session and all its messages
func (s *ChatService) DeleteSession(ctx context.Context, sessionID string) error {
	// Start a transaction to ensure both session and messages are deleted atomically
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First, delete all messages for this session
	deleteMessagesQuery := `DELETE FROM messages WHERE session_id = $1`
	result, err := tx.ExecContext(ctx, deleteMessagesQuery, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session messages: %w", err)
	}

	messagesDeleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get messages deleted count: %w", err)
	}

	// Then, delete the session itself
	deleteSessionQuery := `DELETE FROM sessions WHERE id = $1`
	result, err = tx.ExecContext(ctx, deleteSessionQuery, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	sessionsDeleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get sessions deleted count: %w", err)
	}

	if sessionsDeleted == 0 {
		return fmt.Errorf("session not found")
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Info().
		Str("session_id", sessionID).
		Int64("messages_deleted", messagesDeleted).
		Msg("Session and messages deleted successfully")

	return nil
}
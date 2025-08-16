package services

import (
	"context"
	"fmt"
	"time"

	"chat_ollama/internal/database"
	"chat_ollama/internal/models"
	"chat_ollama/internal/utils"

	"github.com/google/uuid"
)

// ChatService handles chat orchestration and message persistence
type ChatService struct {
	db           *database.DB
	ollamaClient *OllamaClient
	modelManager *ModelManager
	logger       *utils.Logger
}

// NewChatService creates a new chat service
func NewChatService(db *database.DB, ollamaClient *OllamaClient, logger *utils.Logger) *ChatService {
	modelManager := NewModelManager(db, ollamaClient, logger)
	return &ChatService{
		db:           db,
		ollamaClient: ollamaClient,
		modelManager: modelManager,
		logger:       logger.WithComponent("chat_service"),
	}
}

// ProcessChat handles a non-streaming chat request
func (s *ChatService) ProcessChat(ctx context.Context, req models.ChatRequest) (*models.ChatResponse, error) {
	// Validate model if specified
	if req.Model != "" {
		if err := s.modelManager.ValidateModel(ctx, req.Model); err != nil {
			return nil, fmt.Errorf("invalid model: %w", err)
		}
	} else {
		// Use default model if none specified
		defaultModel, err := s.modelManager.GetDefaultModel(ctx)
		if err != nil {
			s.logger.Warn().Err(err).Msg("No default model found, using fallback")
			req.Model = "llama2:7b" // Fallback model
		} else {
			req.Model = defaultModel.Name
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

	s.logger.Info().
		Str("session_id", req.SessionID).
		Str("model", req.Model).
		Int("history_count", len(messages)).
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

	// Send to Ollama
	ollamaResp, err := s.ollamaClient.Chat(ctx, req, messages)
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
		if err := s.modelManager.ValidateModel(ctx, req.Model); err != nil {
			responseChan <- models.StreamResponse{
				Type:      "error",
				SessionID: req.SessionID,
				Error:     fmt.Sprintf("Invalid model: %v", err),
			}
			return err
		}
	} else {
		// Use default model if none specified
		defaultModel, err := s.modelManager.GetDefaultModel(ctx)
		if err != nil {
			s.logger.Warn().Err(err).Msg("No default model found, using fallback")
			req.Model = "llama2:7b" // Fallback model
		} else {
			req.Model = defaultModel.Name
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
	
	// Start streaming from Ollama
	go func() {
		if err := s.ollamaClient.ChatStream(ctx, req, messages, ollamaResponseChan); err != nil {
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
		VALUES (?, ?, ?, ?, ?, ?, ?)
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
		WHERE session_id = ?
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
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM sessions WHERE id = ?)", sessionID).Scan(&exists)
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
		VALUES (?, ?, ?, ?)
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
		SET title = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
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
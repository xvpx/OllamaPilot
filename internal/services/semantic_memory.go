package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"chat_ollama/internal/database"
	"chat_ollama/internal/models"
	"chat_ollama/internal/utils"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// SemanticMemoryService handles semantic memory operations
type SemanticMemoryService struct {
	db               database.Database
	embeddingService *EmbeddingService
	logger           *utils.Logger
	defaultModel     string
}

// MemorySearchResult represents a semantic search result
type MemorySearchResult struct {
	MessageID   string    `json:"message_id"`
	SessionID   string    `json:"session_id"`
	Content     string    `json:"content"`
	Role        string    `json:"role"`
	Similarity  float64   `json:"similarity"`
	CreatedAt   time.Time `json:"created_at"`
	Model       string    `json:"model,omitempty"`
}

// MemorySummary represents a memory summary
type MemorySummary struct {
	ID             string    `json:"id"`
	SessionID      string    `json:"session_id,omitempty"`
	SummaryType    string    `json:"summary_type"`
	Title          string    `json:"title,omitempty"`
	Content        string    `json:"content"`
	RelevanceScore float64   `json:"relevance_score"`
	StartTime      time.Time `json:"start_time,omitempty"`
	EndTime        time.Time `json:"end_time,omitempty"`
	MessageCount   int       `json:"message_count"`
	CreatedAt      time.Time `json:"created_at"`
}

// MemoryGap represents a detected memory gap
type MemoryGap struct {
	ID             string    `json:"id"`
	SessionID      string    `json:"session_id"`
	GapStart       time.Time `json:"gap_start"`
	GapEnd         time.Time `json:"gap_end"`
	ContextSummary string    `json:"context_summary,omitempty"`
	BridgeContent  string    `json:"bridge_content,omitempty"`
	GapType        string    `json:"gap_type"`
	CreatedAt      time.Time `json:"created_at"`
}

// NewSemanticMemoryService creates a new semantic memory service
func NewSemanticMemoryService(db database.Database, embeddingService *EmbeddingService, logger *utils.Logger) *SemanticMemoryService {
	return &SemanticMemoryService{
		db:               db,
		embeddingService: embeddingService,
		logger:           logger.WithComponent("semantic_memory"),
		defaultModel:     "nomic-embed-text", // This will be overridden by config if needed
	}
}

// NewSemanticMemoryServiceWithModel creates a new semantic memory service with a specific embedding model
func NewSemanticMemoryServiceWithModel(db database.Database, embeddingService *EmbeddingService, logger *utils.Logger, embeddingModel string) *SemanticMemoryService {
	return &SemanticMemoryService{
		db:               db,
		embeddingService: embeddingService,
		logger:           logger.WithComponent("semantic_memory"),
		defaultModel:     embeddingModel,
	}
}

// StoreMessageEmbedding generates and stores an embedding for a message
func (s *SemanticMemoryService) StoreMessageEmbedding(ctx context.Context, message models.Message) error {
	// Log the incoming message for debugging with more detail
	s.logger.Info().
		Str("message_id", message.ID).
		Str("session_id", message.SessionID).
		Str("role", message.Role).
		Int("content_length", len(message.Content)).
		Time("created_at", message.CreatedAt).
		Str("content_preview", func() string {
			if len(message.Content) > 50 {
				return message.Content[:50] + "..."
			}
			return message.Content
		}()).
		Msg("DEBUG: Received message for embedding storage")

	// Validate required fields with detailed logging
	if message.SessionID == "" {
		s.logger.Error().
			Str("message_id", message.ID).
			Str("role", message.Role).
			Str("session_id_value", fmt.Sprintf("'%s'", message.SessionID)).
			Int("session_id_length", len(message.SessionID)).
			Msg("ERROR: message.SessionID is empty")
		return fmt.Errorf("message.SessionID cannot be empty")
	}
	if message.Role == "" {
		s.logger.Error().
			Str("message_id", message.ID).
			Str("session_id", message.SessionID).
			Msg("ERROR: message.Role is empty")
		return fmt.Errorf("message.Role cannot be empty")
	}
	if message.Content == "" {
		s.logger.Error().
			Str("message_id", message.ID).
			Str("session_id", message.SessionID).
			Str("role", message.Role).
			Msg("ERROR: message.Content is empty")
		return fmt.Errorf("message.Content cannot be empty")
	}

	// Additional validation to ensure session_id is not nil when passed to database
	sessionID := message.SessionID
	if sessionID == "" {
		s.logger.Error().
			Str("message_id", message.ID).
			Str("role", message.Role).
			Msg("CRITICAL: sessionID variable is empty after assignment")
		return fmt.Errorf("sessionID variable is empty after assignment from message.SessionID")
	}

	// Generate embedding for the message content
	embedding, err := s.embeddingService.GenerateEmbedding(ctx, message.Content, s.defaultModel)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	s.logger.Info().
		Str("message_id", message.ID).
		Str("session_id", message.SessionID).
		Str("role", message.Role).
		Int("content_length", len(message.Content)).
		Int("embedding_dimensions", len(embedding)).
		Msg("Generated embedding, storing in database")

	// Store embedding in database with denormalized message data
	if pgDB, ok := s.db.(*database.PostgresDB); ok {
		// Use the validated sessionID variable instead of message.SessionID
		embeddingID := uuid.New().String()
		
		s.logger.Info().
			Str("embedding_id", embeddingID).
			Str("message_id", message.ID).
			Str("session_id", sessionID).
			Str("role", message.Role).
			Msg("Inserting embedding into database")
		
		err := pgDB.InsertVector(
			"message_embeddings",
			[]string{"id", "message_id", "session_id", "role", "content", "embedding", "model_used", "message_created_at", "created_at"},
			[]interface{}{
				embeddingID,
				message.ID,
				sessionID, // Use validated sessionID variable
				message.Role,
				message.Content,
				pgvector.NewVector(embedding),
				s.defaultModel,
				message.CreatedAt,
				time.Now(),
			},
		)
		
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("embedding_id", embeddingID).
				Str("message_id", message.ID).
				Str("session_id", sessionID).
				Str("role", message.Role).
				Msg("Failed to insert embedding into database")
			return fmt.Errorf("failed to insert embedding into database: %w", err)
		}
		
		s.logger.Info().
			Str("embedding_id", embeddingID).
			Str("message_id", message.ID).
			Str("session_id", sessionID).
			Msg("Successfully stored embedding in database")
		
		return nil
	}

	return fmt.Errorf("vector storage not supported for this database type")
}

// SearchSimilarMessages finds messages similar to the given query
func (s *SemanticMemoryService) SearchSimilarMessages(ctx context.Context, query string, limit int, sessionID string) ([]MemorySearchResult, error) {
	// Generate embedding for the query
	queryEmbedding, err := s.embeddingService.GenerateEmbedding(ctx, query, s.defaultModel)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	_, ok := s.db.(*database.PostgresDB)
	if !ok {
		return nil, fmt.Errorf("semantic search not supported for this database type")
	}

	// Build query with optional session filter - no need to join with messages table anymore
	var sqlQuery string
	var args []interface{}
	
	if sessionID != "" {
		sqlQuery = `
			SELECT message_id, session_id, content, role, message_created_at, model_used,
				   (embedding <=> $1) as distance
			FROM message_embeddings
			WHERE session_id = $2
			ORDER BY embedding <=> $1
			LIMIT $3
		`
		args = []interface{}{pgvector.NewVector(queryEmbedding), sessionID, limit}
	} else {
		sqlQuery = `
			SELECT message_id, session_id, content, role, message_created_at, model_used,
				   (embedding <=> $1) as distance
			FROM message_embeddings
			ORDER BY embedding <=> $1
			LIMIT $2
		`
		args = []interface{}{pgvector.NewVector(queryEmbedding), limit}
	}

	rows, err := s.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute similarity search: %w", err)
	}
	defer rows.Close()

	var results []MemorySearchResult
	for rows.Next() {
		var result MemorySearchResult
		var distance float64
		var model sql.NullString

		err := rows.Scan(
			&result.MessageID,
			&result.SessionID,
			&result.Content,
			&result.Role,
			&result.CreatedAt,
			&model,
			&distance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		if model.Valid {
			result.Model = model.String
		}

		// Convert distance to similarity (1 - distance for cosine distance)
		result.Similarity = 1.0 - distance

		results = append(results, result)
	}

	s.logger.Info().
		Str("query", query).
		Int("results_count", len(results)).
		Str("session_id", sessionID).
		Msg("Semantic search completed")

	return results, nil
}

// CreateMemorySummary creates a summary of a conversation or topic
func (s *SemanticMemoryService) CreateMemorySummary(ctx context.Context, sessionID, summaryType, title, content string, startTime, endTime time.Time, messageCount int) (*MemorySummary, error) {
	// Generate embedding for the summary content
	embedding, err := s.embeddingService.GenerateEmbedding(ctx, content, s.defaultModel)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary embedding: %w", err)
	}

	summary := &MemorySummary{
		ID:           uuid.New().String(),
		SessionID:    sessionID,
		SummaryType:  summaryType,
		Title:        title,
		Content:      content,
		StartTime:    startTime,
		EndTime:      endTime,
		MessageCount: messageCount,
		CreatedAt:    time.Now(),
	}

	// Store summary in database
	if pgDB, ok := s.db.(*database.PostgresDB); ok {
		err = pgDB.InsertVector(
			"memory_summaries",
			[]string{"id", "session_id", "summary_type", "title", "content", "embedding", "start_time", "end_time", "message_count", "created_at"},
			[]interface{}{
				summary.ID,
				summary.SessionID,
				summary.SummaryType,
				summary.Title,
				summary.Content,
				pgvector.NewVector(embedding),
				summary.StartTime,
				summary.EndTime,
				summary.MessageCount,
				summary.CreatedAt,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to store memory summary: %w", err)
		}
	} else {
		return nil, fmt.Errorf("memory summaries not supported for this database type")
	}

	s.logger.Info().
		Str("summary_id", summary.ID).
		Str("session_id", sessionID).
		Str("summary_type", summaryType).
		Msg("Memory summary created")

	return summary, nil
}

// DetectMemoryGaps identifies gaps in conversation context
func (s *SemanticMemoryService) DetectMemoryGaps(ctx context.Context, sessionID string, timeThreshold time.Duration) ([]MemoryGap, error) {
	// Query messages in chronological order
	query := `
		SELECT id, created_at, content
		FROM messages
		WHERE session_id = $1
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages for gap detection: %w", err)
	}
	defer rows.Close()

	var messages []struct {
		ID        string
		CreatedAt time.Time
		Content   string
	}

	for rows.Next() {
		var msg struct {
			ID        string
			CreatedAt time.Time
			Content   string
		}
		if err := rows.Scan(&msg.ID, &msg.CreatedAt, &msg.Content); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	var gaps []MemoryGap

	// Detect temporal gaps
	for i := 1; i < len(messages); i++ {
		timeDiff := messages[i].CreatedAt.Sub(messages[i-1].CreatedAt)
		if timeDiff > timeThreshold {
			gap := MemoryGap{
				ID:        uuid.New().String(),
				SessionID: sessionID,
				GapStart:  messages[i-1].CreatedAt,
				GapEnd:    messages[i].CreatedAt,
				GapType:   "temporal",
				CreatedAt: time.Now(),
			}

			// Generate context summary for the gap
			contextBefore := messages[i-1].Content
			contextAfter := messages[i].Content
			gap.ContextSummary = fmt.Sprintf("Gap of %v between: '%s' and '%s'",
				timeDiff, truncateText(contextBefore, 100), truncateText(contextAfter, 100))

			gaps = append(gaps, gap)
		}
	}

	// Store detected gaps
	for _, gap := range gaps {
		_, err := s.db.Exec(`
			INSERT INTO memory_gaps (id, session_id, gap_start, gap_end, context_summary, gap_type, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, gap.ID, gap.SessionID, gap.GapStart, gap.GapEnd, gap.ContextSummary, gap.GapType, gap.CreatedAt)
		
		if err != nil {
			s.logger.Error().Err(err).Str("gap_id", gap.ID).Msg("Failed to store memory gap")
		}
	}

	s.logger.Info().
		Str("session_id", sessionID).
		Int("gaps_detected", len(gaps)).
		Msg("Memory gap detection completed")

	return gaps, nil
}

// GetRelevantContext retrieves relevant context for a query using semantic search
func (s *SemanticMemoryService) GetRelevantContext(ctx context.Context, query string, sessionID string, maxResults int) (string, error) {
	// Search for similar messages
	results, err := s.SearchSimilarMessages(ctx, query, maxResults, sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to search similar messages: %w", err)
	}

	if len(results) == 0 {
		return "", nil
	}

	// Build context from search results
	var contextParts []string
	for _, result := range results {
		// Debug: Include ALL results regardless of similarity
		contextParts = append(contextParts, fmt.Sprintf("[%s] %s", result.Role, result.Content))
	}

	return strings.Join(contextParts, "\n"), nil
}

// ProcessMessageForSemanticMemory processes a new message for semantic memory features
func (s *SemanticMemoryService) ProcessMessageForSemanticMemory(ctx context.Context, message models.Message) error {
	// Store message embedding
	if err := s.StoreMessageEmbedding(ctx, message); err != nil {
		s.logger.Error().Err(err).Str("message_id", message.ID).Msg("Failed to store message embedding")
		// Don't return error as this shouldn't block the main chat flow
	}

	// Detect memory gaps (run asynchronously)
	go func() {
		gaps, err := s.DetectMemoryGaps(context.Background(), message.SessionID, 1*time.Hour)
		if err != nil {
			s.logger.Error().Err(err).Str("session_id", message.SessionID).Msg("Failed to detect memory gaps")
		} else if len(gaps) > 0 {
			s.logger.Info().Str("session_id", message.SessionID).Int("gaps", len(gaps)).Msg("Memory gaps detected")
		}
	}()

	return nil
}

// CleanupObsoleteEmbeddings removes embeddings that are no longer useful
// This is a maintenance function that can be called periodically
func (s *SemanticMemoryService) CleanupObsoleteEmbeddings(ctx context.Context, options CleanupOptions) (*CleanupResult, error) {
	result := &CleanupResult{}
	
	// Build cleanup query based on options
	var conditions []string
	var args []interface{}
	argIndex := 1
	
	if options.OlderThanDays > 0 {
		conditions = append(conditions, fmt.Sprintf("message_created_at < NOW() - INTERVAL '%d days'", options.OlderThanDays))
	}
	
	if options.SessionID != "" {
		conditions = append(conditions, fmt.Sprintf("session_id = $%d", argIndex))
		args = append(args, options.SessionID)
		argIndex++
	}
	
	if len(options.ExcludeRoles) > 0 {
		placeholders := make([]string, len(options.ExcludeRoles))
		for i, role := range options.ExcludeRoles {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, role)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("role NOT IN (%s)", strings.Join(placeholders, ",")))
	}
	
	if options.MinSimilarityThreshold > 0 {
		// This is more complex - we'd need to compare against recent queries
		// For now, we'll skip this advanced feature
		s.logger.Warn().Msg("MinSimilarityThreshold cleanup not yet implemented")
	}
	
	if len(conditions) == 0 {
		return result, fmt.Errorf("no cleanup conditions specified - this would delete all embeddings")
	}
	
	// Count embeddings to be deleted
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM message_embeddings WHERE %s", strings.Join(conditions, " AND "))
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&result.EmbeddingsDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to count embeddings for cleanup: %w", err)
	}
	
	if options.DryRun {
		result.DryRun = true
		s.logger.Info().
			Int64("would_delete", result.EmbeddingsDeleted).
			Msg("Dry run: embeddings that would be deleted")
		return result, nil
	}
	
	// Perform the actual deletion
	deleteQuery := fmt.Sprintf("DELETE FROM message_embeddings WHERE %s", strings.Join(conditions, " AND "))
	deleteResult, err := s.db.ExecContext(ctx, deleteQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to delete obsolete embeddings: %w", err)
	}
	
	actualDeleted, err := deleteResult.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get deleted count: %w", err)
	}
	
	result.EmbeddingsDeleted = actualDeleted
	
	s.logger.Info().
		Int64("embeddings_deleted", result.EmbeddingsDeleted).
		Msg("Cleanup completed")
	
	return result, nil
}

// GetEmbeddingStats returns statistics about stored embeddings
func (s *SemanticMemoryService) GetEmbeddingStats(ctx context.Context) (*EmbeddingStats, error) {
	stats := &EmbeddingStats{}
	
	// Total embeddings
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM message_embeddings").Scan(&stats.TotalEmbeddings)
	if err != nil {
		return nil, fmt.Errorf("failed to get total embeddings count: %w", err)
	}
	
	// Embeddings by role
	roleQuery := `
		SELECT role, COUNT(*)
		FROM message_embeddings
		GROUP BY role
	`
	rows, err := s.db.QueryContext(ctx, roleQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get embeddings by role: %w", err)
	}
	defer rows.Close()
	
	stats.ByRole = make(map[string]int64)
	for rows.Next() {
		var role string
		var count int64
		if err := rows.Scan(&role, &count); err != nil {
			return nil, fmt.Errorf("failed to scan role stats: %w", err)
		}
		stats.ByRole[role] = count
	}
	
	// Oldest and newest embeddings
	err = s.db.QueryRowContext(ctx, "SELECT MIN(message_created_at), MAX(message_created_at) FROM message_embeddings").
		Scan(&stats.OldestEmbedding, &stats.NewestEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding date range: %w", err)
	}
	
	// Unique sessions with embeddings
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT session_id) FROM message_embeddings").Scan(&stats.UniqueSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to get unique sessions count: %w", err)
	}
	
	return stats, nil
}

// CleanupOptions defines options for embedding cleanup
type CleanupOptions struct {
	OlderThanDays           int      // Delete embeddings older than this many days
	SessionID               string   // Only clean embeddings from this session (empty = all sessions)
	ExcludeRoles            []string // Don't delete embeddings with these roles
	MinSimilarityThreshold  float64  // Delete embeddings that haven't been similar to recent queries (not implemented yet)
	DryRun                  bool     // If true, only count what would be deleted
}

// CleanupResult contains the results of a cleanup operation
type CleanupResult struct {
	EmbeddingsDeleted int64 `json:"embeddings_deleted"`
	DryRun            bool  `json:"dry_run"`
}

// EmbeddingStats contains statistics about stored embeddings
type EmbeddingStats struct {
	TotalEmbeddings  int64            `json:"total_embeddings"`
	ByRole           map[string]int64 `json:"by_role"`
	OldestEmbedding  time.Time        `json:"oldest_embedding"`
	NewestEmbedding  time.Time        `json:"newest_embedding"`
	UniqueSessions   int64            `json:"unique_sessions"`
}
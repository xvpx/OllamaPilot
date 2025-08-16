package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"chat_ollama/internal/models"
	"chat_ollama/internal/utils"
)

// OllamaClient handles communication with Ollama API
type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *utils.Logger
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(host string, timeout time.Duration, logger *utils.Logger) *OllamaClient {
	baseURL := fmt.Sprintf("http://%s", host)
	if !strings.HasPrefix(host, "http") {
		baseURL = fmt.Sprintf("http://%s", host)
	}

	return &OllamaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger.WithComponent("ollama_client"),
	}
}

// OllamaMessage represents a message in Ollama format
type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaChatRequest represents a chat request to Ollama
type OllamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// OllamaChatResponse represents a chat response from Ollama
type OllamaChatResponse struct {
	Model     string        `json:"model"`
	CreatedAt time.Time     `json:"created_at"`
	Message   OllamaMessage `json:"message"`
	Done      bool          `json:"done"`
	TotalDuration int64     `json:"total_duration,omitempty"`
	LoadDuration  int64     `json:"load_duration,omitempty"`
	PromptEvalCount int     `json:"prompt_eval_count,omitempty"`
	EvalCount     int       `json:"eval_count,omitempty"`
}

// HealthCheck checks if Ollama is available
func (c *OllamaClient) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ollama health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama health check returned status %d", resp.StatusCode)
	}

	c.logger.Debug().Msg("Ollama health check passed")
	return nil
}

// Chat sends a chat request to Ollama (non-streaming)
func (c *OllamaClient) Chat(ctx context.Context, req models.ChatRequest, messages []models.Message) (*OllamaChatResponse, error) {
	// Convert messages to Ollama format
	ollamaMessages := c.convertMessages(messages)
	
	// Add the current user message
	ollamaMessages = append(ollamaMessages, OllamaMessage{
		Role:    "user",
		Content: req.Message,
	})

	ollamaReq := OllamaChatRequest{
		Model:    req.Model,
		Messages: ollamaMessages,
		Stream:   false,
		Options:  req.Options,
	}

	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	c.logger.Debug().
		Str("model", req.Model).
		Int("message_count", len(ollamaMessages)).
		Msg("Sending chat request to Ollama")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug().
		Str("model", ollamaResp.Model).
		Bool("done", ollamaResp.Done).
		Int("eval_count", ollamaResp.EvalCount).
		Msg("Received response from Ollama")

	return &ollamaResp, nil
}

// ChatStream sends a streaming chat request to Ollama
func (c *OllamaClient) ChatStream(ctx context.Context, req models.ChatRequest, messages []models.Message, responseChan chan<- models.StreamResponse) error {
	defer close(responseChan)

	// Convert messages to Ollama format
	ollamaMessages := c.convertMessages(messages)
	
	// Add the current user message
	ollamaMessages = append(ollamaMessages, OllamaMessage{
		Role:    "user",
		Content: req.Message,
	})

	ollamaReq := OllamaChatRequest{
		Model:    req.Model,
		Messages: ollamaMessages,
		Stream:   true,
		Options:  req.Options,
	}

	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		responseChan <- models.StreamResponse{
			Type:      "error",
			SessionID: req.SessionID,
			Error:     fmt.Sprintf("Failed to marshal request: %v", err),
		}
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewBuffer(reqBody))
	if err != nil {
		responseChan <- models.StreamResponse{
			Type:      "error",
			SessionID: req.SessionID,
			Error:     fmt.Sprintf("Failed to create request: %v", err),
		}
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	c.logger.Debug().
		Str("model", req.Model).
		Str("session_id", req.SessionID).
		Int("message_count", len(ollamaMessages)).
		Msg("Starting streaming chat request to Ollama")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		responseChan <- models.StreamResponse{
			Type:      "error",
			SessionID: req.SessionID,
			Error:     fmt.Sprintf("Failed to send request to Ollama: %v", err),
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		responseChan <- models.StreamResponse{
			Type:      "error",
			SessionID: req.SessionID,
			Error:     fmt.Sprintf("Ollama returned status %d: %s", resp.StatusCode, string(body)),
		}
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	// Read streaming response
	decoder := json.NewDecoder(resp.Body)
	var totalTokens int

	for {
		select {
		case <-ctx.Done():
			responseChan <- models.StreamResponse{
				Type:      "error",
				SessionID: req.SessionID,
				Error:     "Request cancelled",
			}
			return ctx.Err()
		default:
		}

		var ollamaResp OllamaChatResponse
		if err := decoder.Decode(&ollamaResp); err != nil {
			if err == io.EOF {
				break
			}
			responseChan <- models.StreamResponse{
				Type:      "error",
				SessionID: req.SessionID,
				Error:     fmt.Sprintf("Failed to decode streaming response: %v", err),
			}
			return err
		}

		if ollamaResp.Message.Content != "" {
			responseChan <- models.StreamResponse{
				Type:      "token",
				Content:   ollamaResp.Message.Content,
				SessionID: req.SessionID,
			}
		}

		if ollamaResp.Done {
			totalTokens = ollamaResp.EvalCount
			break
		}
	}

	// Send completion message
	responseChan <- models.StreamResponse{
		Type:      "done",
		SessionID: req.SessionID,
		Metadata: map[string]interface{}{
			"total_tokens": totalTokens,
			"model":        req.Model,
		},
	}

	c.logger.Debug().
		Str("session_id", req.SessionID).
		Int("total_tokens", totalTokens).
		Msg("Completed streaming chat request")

	return nil
}

// convertMessages converts internal message format to Ollama format
func (c *OllamaClient) convertMessages(messages []models.Message) []OllamaMessage {
	ollamaMessages := make([]OllamaMessage, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = OllamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return ollamaMessages
}

// GetModels retrieves available models from Ollama
func (c *OllamaClient) GetModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, len(result.Models))
	for i, model := range result.Models {
		models[i] = model.Name
	}

	return models, nil
}

// GetModelsWithInfo retrieves models with detailed information from Ollama
func (c *OllamaClient) GetModelsWithInfo(ctx context.Context) ([]models.OllamaModelDetailedInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result struct {
		Models []struct {
			Name    string `json:"name"`
			Size    int64  `json:"size"`
			Details struct {
				Family           string `json:"family"`
				Format           string `json:"format"`
				ParameterSize    string `json:"parameter_size"`
				QuantizationLevel string `json:"quantization_level"`
			} `json:"details"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	modelsInfo := make([]models.OllamaModelDetailedInfo, len(result.Models))
	for i, model := range result.Models {
		modelsInfo[i] = models.OllamaModelDetailedInfo{
			Name:         model.Name,
			Size:         model.Size,
			Family:       model.Details.Family,
			Format:       model.Details.Format,
			Parameters:   model.Details.ParameterSize,
			Quantization: model.Details.QuantizationLevel,
		}
	}

	return modelsInfo, nil
}

// PullModel downloads a model from Ollama
func (c *OllamaClient) PullModel(ctx context.Context, modelName string, progressChan chan<- models.ModelDownloadProgress) error {
	defer close(progressChan)

	pullReq := struct {
		Name   string `json:"name"`
		Stream bool   `json:"stream"`
	}{
		Name:   modelName,
		Stream: true,
	}

	reqBody, err := json.Marshal(pullReq)
	if err != nil {
		progressChan <- models.ModelDownloadProgress{
			ModelName: modelName,
			Status:    "error",
			Error:     fmt.Sprintf("Failed to marshal request: %v", err),
		}
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/pull", bytes.NewBuffer(reqBody))
	if err != nil {
		progressChan <- models.ModelDownloadProgress{
			ModelName: modelName,
			Status:    "error",
			Error:     fmt.Sprintf("Failed to create request: %v", err),
		}
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	c.logger.Info().
		Str("model", modelName).
		Msg("Starting model download from Ollama")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		progressChan <- models.ModelDownloadProgress{
			ModelName: modelName,
			Status:    "error",
			Error:     fmt.Sprintf("Failed to send request to Ollama: %v", err),
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		progressChan <- models.ModelDownloadProgress{
			ModelName: modelName,
			Status:    "error",
			Error:     fmt.Sprintf("Ollama returned status %d: %s", resp.StatusCode, string(body)),
		}
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	// Read streaming response
	decoder := json.NewDecoder(resp.Body)

	for {
		select {
		case <-ctx.Done():
			progressChan <- models.ModelDownloadProgress{
				ModelName: modelName,
				Status:    "error",
				Error:     "Download cancelled",
			}
			return ctx.Err()
		default:
		}

		var pullResp struct {
			Status    string `json:"status"`
			Digest    string `json:"digest,omitempty"`
			Total     int64  `json:"total,omitempty"`
			Completed int64  `json:"completed,omitempty"`
		}

		if err := decoder.Decode(&pullResp); err != nil {
			if err == io.EOF {
				break
			}
			progressChan <- models.ModelDownloadProgress{
				ModelName: modelName,
				Status:    "error",
				Error:     fmt.Sprintf("Failed to decode streaming response: %v", err),
			}
			return err
		}

		progress := models.ModelDownloadProgress{
			ModelName: modelName,
			Status:    pullResp.Status,
			Digest:    pullResp.Digest,
			Total:     pullResp.Total,
			Completed: pullResp.Completed,
		}

		if pullResp.Total > 0 {
			progress.Percentage = float64(pullResp.Completed) / float64(pullResp.Total) * 100
		}

		progressChan <- progress

		// Check if download is complete
		if pullResp.Status == "success" || strings.Contains(pullResp.Status, "success") {
			break
		}
	}

	c.logger.Info().
		Str("model", modelName).
		Msg("Model download completed")

	return nil
}

// GetModelInfo retrieves detailed information about a model
func (c *OllamaClient) GetModelInfo(ctx context.Context, modelName string) (*models.OllamaModelInfo, error) {
	showReq := struct {
		Name string `json:"name"`
	}{
		Name: modelName,
	}

	reqBody, err := json.Marshal(showReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/show", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get model info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var modelInfo models.OllamaModelInfo
	if err := json.NewDecoder(resp.Body).Decode(&modelInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &modelInfo, nil
}

// GetLibraryModels fetches available models from Ollama's public library
func (c *OllamaClient) GetLibraryModels(ctx context.Context) ([]string, error) {
	c.logger.Info().Msg("Fetching models from Ollama library")
	
	// First, get the list of model families from the library page
	req, err := http.NewRequestWithContext(ctx, "GET", "https://ollama.com/library", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create library request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch library page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("library page returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read library page: %w", err)
	}

	// Extract model family names from the HTML
	content := string(body)
	modelFamilies := c.extractModelFamilies(content)
	
	c.logger.Info().Int("families_found", len(modelFamilies)).Msg("Found model families")

	// For each model family, get the available variants
	var allModels []string
	for _, family := range modelFamilies {
		variants, err := c.getModelVariants(ctx, family)
		if err != nil {
			c.logger.Warn().Err(err).Str("family", family).Msg("Failed to get variants for model family")
			// Add the base family name if we can't get variants
			allModels = append(allModels, family+":latest")
			continue
		}
		allModels = append(allModels, variants...)
	}

	c.logger.Info().Int("total_models", len(allModels)).Msg("Successfully fetched models from library")
	return allModels, nil
}

// extractModelFamilies extracts model family names from the library HTML page
func (c *OllamaClient) extractModelFamilies(content string) []string {
	var families []string
	
	// Look for href="/library/modelname" patterns
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, `href="/library/`) {
			// Extract the model name from href="/library/modelname"
			start := strings.Index(line, `href="/library/`)
			if start == -1 {
				continue
			}
			start += len(`href="/library/`)
			end := strings.Index(line[start:], `"`)
			if end == -1 {
				continue
			}
			modelName := line[start : start+end]
			if modelName != "" && !strings.Contains(modelName, "/") {
				families = append(families, modelName)
			}
		}
	}
	
	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueFamilies []string
	for _, family := range families {
		if !seen[family] {
			seen[family] = true
			uniqueFamilies = append(uniqueFamilies, family)
		}
	}
	
	return uniqueFamilies
}

// getModelVariants fetches the available variants for a specific model family
func (c *OllamaClient) getModelVariants(ctx context.Context, family string) ([]string, error) {
	url := fmt.Sprintf("https://ollama.com/library/%s", family)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", family, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s page: %w", family, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s page returned status %d", family, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s page: %w", family, err)
	}

	content := string(body)
	variants := c.extractModelVariants(content, family)
	
	// If no variants found, add the default latest tag
	if len(variants) == 0 {
		variants = append(variants, family+":latest")
	}
	
	return variants, nil
}

// extractModelVariants extracts model variants from a model family page
func (c *OllamaClient) extractModelVariants(content, family string) []string {
	var variants []string
	
	// Look for patterns like "modelname:tag"
	lines := strings.Split(content, "\n")
	pattern := family + ":"
	
	seen := make(map[string]bool)
	for _, line := range lines {
		if strings.Contains(line, pattern) {
			// Find all occurrences of the pattern in the line
			start := 0
			for {
				idx := strings.Index(line[start:], pattern)
				if idx == -1 {
					break
				}
				
				fullStart := start + idx
				// Find the end of the model name (until space, quote, or HTML tag)
				end := fullStart + len(pattern)
				for end < len(line) && line[end] != ' ' && line[end] != '"' && line[end] != '<' && line[end] != '>' {
					end++
				}
				
				if end > fullStart+len(pattern) {
					variant := line[fullStart:end]
					// Clean up any trailing punctuation or HTML
					variant = strings.TrimRight(variant, ".,;:")
					if variant != "" && !seen[variant] {
						seen[variant] = true
						variants = append(variants, variant)
					}
				}
				
				start = fullStart + 1
			}
		}
	}
	
	return variants
}

// DeleteModel removes a model from Ollama
func (c *OllamaClient) DeleteModel(ctx context.Context, modelName string) error {
	deleteReq := struct {
		Name string `json:"name"`
	}{
		Name: modelName,
	}

	reqBody, err := json.Marshal(deleteReq)
	if err != nil {
		return fmt.Errorf("failed to marshal delete request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+"/api/delete", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	c.logger.Info().
		Str("model", modelName).
		Msg("Deleting model from Ollama")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send delete request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.Info().
		Str("model", modelName).
		Msg("Model deleted from Ollama successfully")

	return nil
}
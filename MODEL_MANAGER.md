# Model Manager Documentation

## Overview

The Model Manager is a comprehensive system for managing language models in the Chat Ollama application. It provides full CRUD operations, configuration management, usage tracking, and a user-friendly web interface.

## Features

### Core Functionality
- **Model Discovery**: Automatically sync with Ollama to discover available models
- **Model Management**: Enable/disable, set default models, and manage metadata
- **Configuration Management**: Per-model configuration with temperature, top-p, context length, etc.
- **Usage Tracking**: Track model usage statistics and last used timestamps
- **Web Interface**: Tabbed UI for easy model management

### Database Schema
- **models**: Core model metadata and status
- **model_configs**: Per-model configuration settings
- **model_usage_stats**: Aggregated usage statistics (view)

## API Endpoints

### Model Management
```
GET    /v1/models                    # List all models
GET    /v1/models?available=true     # List only available models
GET    /v1/models/{id}               # Get model details
PUT    /v1/models/{id}               # Update model metadata
DELETE /v1/models/{id}               # Mark model as removed
POST   /v1/models/sync               # Sync with Ollama
```

### Model Configuration
```
GET    /v1/models/{id}/config        # Get model configuration
PUT    /v1/models/{id}/config        # Update model configuration
```

### Model Operations
```
POST   /v1/models/{id}/default       # Set as default model
GET    /v1/models/{id}/stats         # Get usage statistics
```

## Usage Examples

### 1. Sync Models with Ollama
```bash
curl -X POST http://localhost:8081/v1/models/sync
```

### 2. List Available Models
```bash
curl -X GET http://localhost:8081/v1/models?available=true
```

### 3. Get Model Details
```bash
curl -X GET http://localhost:8081/v1/models/{model-id}
```

### 4. Update Model Configuration
```bash
curl -X PUT http://localhost:8081/v1/models/{model-id}/config \
  -H "Content-Type: application/json" \
  -d '{
    "temperature": 0.7,
    "top_p": 0.9,
    "context_length": 4096,
    "system_prompt": "You are a helpful assistant."
  }'
```

### 5. Set Default Model
```bash
curl -X POST http://localhost:8081/v1/models/{model-id}/default
```

### 6. Enable/Disable Model
```bash
curl -X PUT http://localhost:8081/v1/models/{model-id} \
  -H "Content-Type: application/json" \
  -d '{"is_enabled": false}'
```

## Web Interface

### Navigation
The web interface includes a tabbed sidebar with:
- **💬 Chats**: Chat session management
- **🤖 Models**: Model management interface

### Model Management Features
- **Model List**: View all models with status indicators
- **Sync Button**: Manually sync with Ollama
- **Model Actions**: Set default, configure, enable/disable
- **Configuration Panel**: View and edit model parameters
- **Status Indicators**: Available, error, removed states
- **Default Badge**: Visual indicator for default model

### Model Status Types
- **Available**: Model is ready for use
- **Error**: Model has issues
- **Removed**: Model no longer available in Ollama

## Database Schema Details

### Models Table
```sql
CREATE TABLE models (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT DEFAULT '',
    size INTEGER DEFAULT 0,
    family TEXT DEFAULT '',
    format TEXT DEFAULT '',
    parameters TEXT DEFAULT '',
    quantization TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'available',
    is_default BOOLEAN DEFAULT FALSE,
    is_enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME
);
```

### Model Configs Table
```sql
CREATE TABLE model_configs (
    id TEXT PRIMARY KEY,
    model_id TEXT NOT NULL,
    temperature REAL,
    top_p REAL,
    top_k INTEGER,
    repeat_penalty REAL,
    context_length INTEGER,
    max_tokens INTEGER,
    system_prompt TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE
);
```

### Usage Statistics View
```sql
CREATE VIEW model_usage_stats AS
SELECT 
    m.id as model_id,
    m.name,
    COUNT(msg.id) as total_messages,
    COALESCE(SUM(msg.tokens_used), 0) as total_tokens,
    MAX(msg.created_at) as last_used_at
FROM models m
LEFT JOIN messages msg ON m.name = msg.model
GROUP BY m.id, m.name;
```

## Integration with Chat Service

The model manager is integrated with the chat service to provide:

### Model Validation
- Validates model availability before processing chat requests
- Automatically uses default model if none specified
- Provides clear error messages for invalid models

### Usage Tracking
- Automatically updates `last_used_at` when models are used
- Tracks token usage for statistics
- Maintains message counts per model

## Configuration Options

### Default Model Configuration
```go
ModelConfig{
    Temperature:   0.7,
    TopP:          0.9,
    TopK:          40,
    RepeatPenalty: 1.1,
    ContextLength: 4096,
    MaxTokens:     2048,
    SystemPrompt:  "",
}
```

### Model Status Values
- `available`: Model is ready for use
- `downloading`: Model is being downloaded (future feature)
- `installing`: Model is being installed (future feature)
- `error`: Model has errors
- `removed`: Model is no longer available

## Error Handling

### API Error Responses
All endpoints return structured error responses:
```json
{
  "type": "validation_error",
  "title": "Validation Error",
  "status": 400,
  "detail": "Model not found",
  "instance": "/v1/models/invalid-id",
  "timestamp": "2025-08-16T13:26:00Z"
}
```

### Common Error Scenarios
- **Model Not Found**: 404 when accessing non-existent models
- **Invalid Model Status**: 400 when trying to use unavailable models
- **Configuration Errors**: 400 for invalid configuration parameters
- **Sync Failures**: 500 when Ollama is unreachable

## Testing

### Manual Testing Steps

1. **Start the Application**
   ```bash
   go build -o chat_ollama ./cmd/api
   ./chat_ollama
   ```

2. **Test Model Sync**
   ```bash
   curl -X POST http://localhost:8081/v1/models/sync
   ```

3. **Verify Models List**
   ```bash
   curl -X GET http://localhost:8081/v1/models
   ```

4. **Test Web Interface**
   - Open http://localhost:8081
   - Click on "🤖 Models" tab
   - Click "🔄 Sync" button
   - Verify models appear in the list

5. **Test Model Operations**
   - Set a model as default
   - View model configuration
   - Enable/disable models

### Automated Testing
```bash
# Test health endpoint
curl -X GET http://localhost:8081/health

# Test model sync
curl -X POST http://localhost:8081/v1/models/sync

# Test model listing
curl -X GET http://localhost:8081/v1/models?available=true

# Test chat with model validation
curl -X POST http://localhost:8081/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Hello",
    "session_id": "test-123",
    "model": "llama3.2:1b",
    "stream": false
  }'
```

## Future Enhancements

### Planned Features
- **Model Installation**: Direct model installation from Ollama library
- **Model Removal**: Remove models from Ollama
- **Advanced Configuration**: Custom model parameters and fine-tuning
- **Model Templates**: Predefined model configurations
- **Batch Operations**: Bulk model management operations
- **Model Metrics**: Detailed performance and usage analytics

### API Extensions
- `POST /v1/models/install` - Install new models
- `DELETE /v1/models/{id}/remove` - Remove from Ollama
- `GET /v1/models/templates` - Get configuration templates
- `POST /v1/models/batch` - Batch operations

## Troubleshooting

### Common Issues

1. **Models Not Appearing**
   - Ensure Ollama is running and accessible
   - Check Ollama host configuration
   - Run model sync manually

2. **Default Model Not Working**
   - Verify model is enabled and available
   - Check model status in database
   - Ensure model exists in Ollama

3. **Configuration Not Saving**
   - Verify model ID is correct
   - Check request payload format
   - Review server logs for errors

4. **Web Interface Issues**
   - Clear browser cache
   - Check browser console for errors
   - Verify API endpoints are accessible

### Debug Commands
```bash
# Check Ollama connectivity
curl -X GET http://localhost:11434/api/tags

# Check database
psql -h localhost -U postgres -d ollamapilot -c "SELECT * FROM models;"

# Check server logs
tail -f ./logs/app.log
```

## Security Considerations

- Model configurations are stored in the database
- No authentication required for model management (add as needed)
- Input validation on all configuration parameters
- SQL injection protection through parameterized queries
- XSS protection in web interface through HTML escaping

## Performance Notes

- Model sync operations may take time with many models
- Database indexes optimize model queries
- Frontend uses efficient DOM updates for model lists
- Lazy loading of model configurations
- Caching of model metadata for better performance
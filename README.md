# Chat Ollama MVP - Go API Server

A modular, Dockerized Ollama frontend with real-time streaming chat functionality, built with Go and SQLite.

## 🏗️ Architecture

This implementation follows the architectural specification in `ARCHITECTURE.md` and provides:

- **HTTP Server**: Chi router with middleware for logging, CORS, and error handling
- **Database**: SQLite with automated migrations
- **Configuration**: Environment variable-based configuration
- **Logging**: Structured logging with zerolog
- **Error Handling**: RFC 7807 compliant error responses
- **Health Checks**: Kubernetes-ready health endpoints
- **Docker**: Production-ready containerization

## 📁 Project Structure

```
chat_ollama/
├── cmd/api/main.go                 # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/               # HTTP request handlers
│   │   ├── middleware/             # HTTP middleware
│   │   └── router.go               # Route configuration
│   ├── config/                     # Configuration management
│   ├── database/                   # Database connection & migrations
│   ├── models/                     # Data models
│   └── utils/                      # Utilities (errors, logging, responses)
├── migrations/                     # SQL migration files
├── docker-compose.yml              # Multi-container setup
├── Dockerfile                      # API service container
├── Makefile                        # Build automation
└── .env.example                    # Environment variables template
```

## 🚀 Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose (for containerized deployment)

### Local Development

1. **Install Go dependencies:**
   ```bash
   make deps
   ```

2. **Set up environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Create data directory:**
   ```bash
   make setup
   ```

4. **Build and run:**
   ```bash
   make run
   ```

The server will start on `http://localhost:8080`

### Docker Deployment

1. **Run with Docker Compose:**
   ```bash
   make docker-run
   ```

2. **Run in background:**
   ```bash
   make docker-up
   ```

3. **View logs:**
   ```bash
   make docker-logs
   ```

4. **Stop services:**
   ```bash
   make docker-down
   ```

## 🔍 API Endpoints

### Health Checks

- `GET /health` - Comprehensive health check with service status
- `GET /ready` - Kubernetes readiness probe
- `GET /live` - Kubernetes liveness probe
- `GET /ping` - Simple heartbeat endpoint

### Chat API (v1)

- `POST /v1/chat` - Send chat message (streaming/non-streaming)
- `GET /v1/sessions` - List chat sessions
- `GET /v1/sessions/{id}/messages` - Get session message history

### Model Management API (v1)

- `GET /v1/models` - List all models
- `GET /v1/models?available=true` - List only available models
- `GET /v1/models/{id}` - Get model details with configuration and stats
- `PUT /v1/models/{id}` - Update model metadata (enable/disable, set default)
- `DELETE /v1/models/{id}` - Mark model as removed
- `POST /v1/models/sync` - Sync models with Ollama
- `GET /v1/models/{id}/config` - Get model configuration
- `PUT /v1/models/{id}/config` - Update model configuration
- `POST /v1/models/{id}/default` - Set model as default
- `GET /v1/models/{id}/stats` - Get model usage statistics

### Example Health Check Response

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "services": {
    "database": "healthy",
    "ollama": "healthy"
  },
  "metadata": {
    "sqlite_version": "3.45.0"
  }
}
```

### Example Chat Request

```json
{
  "message": "Hello, how are you?",
  "session_id": "uuid-v4-string",
  "model": "llama2:7b",
  "stream": true,
  "options": {
    "temperature": 0.7,
    "max_tokens": 2048
  }
}
```

## 🛠️ Development

### Available Make Commands

```bash
make help           # Show all available commands
make build          # Build the application
make run            # Build and run locally
make dev            # Run with hot reload (requires air)
make test           # Run tests
make clean          # Clean build artifacts
make fmt            # Format code
make lint           # Lint code (requires golangci-lint)
make security       # Security scan (requires gosec)
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `HOST` | `0.0.0.0` | Server host |
| `ENV` | `development` | Environment (development/production) |
| `DB_PATH` | `./data/chat.db` | SQLite database path |
| `OLLAMA_HOST` | `localhost:11434` | Ollama service host |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `LOG_FORMAT` | `json` | Log format (json/console) |

## 🗄️ Database

The application uses SQLite with automated migrations:

- **Sessions**: Chat session metadata
- **Messages**: Individual chat messages with role, content, and tokens
- **Models**: Model metadata, status, and configuration
- **Model Configs**: Per-model configuration settings
- **Migrations**: Automatic schema versioning

### Database Schema

The application includes comprehensive model management with:
- Model discovery and synchronization with Ollama
- Per-model configuration (temperature, top-p, context length, etc.)
- Usage tracking and statistics
- Default model management

### Manual Migration

```bash
make migrate
```

## 🐳 Docker

### Build Image

```bash
make docker-build
```

### Production Deployment

The included `docker-compose.yml` provides:

- **API Service**: Go application with health checks
- **Ollama Service**: LLM runtime with model storage
- **Volumes**: Persistent data and model storage
- **Network**: Isolated container network

## 🔧 Configuration

### Development

Copy `.env.example` to `.env` and customize:

```bash
cp .env.example .env
```

### Production

Set environment variables in your deployment system or Docker Compose.

## 📊 Monitoring

### Health Endpoints

- `/health` - Full health check with dependency status
- `/ready` - Readiness for traffic (database connectivity)
- `/live` - Liveness check (application responsiveness)

### Logging

Structured JSON logging with:
- Request/response logging
- Database query logging
- Error tracking with stack traces
- Performance metrics

## 🚧 Current Implementation Status

✅ **Completed:**
- HTTP server with chi router
- SQLite database with migrations
- Configuration management
- Structured logging and error handling
- Health check endpoints
- Middleware (CORS, logging, recovery)
- Docker containerization
- **Full chat functionality with Ollama integration**
- **Streaming and non-streaming chat support**
- **Message persistence and session management**
- **Comprehensive model management system**
- **Web UI with model management interface**

### 🤖 Model Manager Features

✅ **Model Management:**
- Automatic model discovery and sync with Ollama
- Model enable/disable functionality
- Default model configuration
- Model status tracking (available, error, removed)
- Usage statistics and last-used tracking

✅ **Configuration Management:**
- Per-model configuration settings
- Temperature, top-p, top-k, repeat penalty controls
- Context length and max tokens configuration
- Custom system prompts
- Real-time configuration updates

✅ **Web Interface:**
- Tabbed interface for chat and model management
- Visual model status indicators
- One-click model operations (set default, enable/disable)
- Configuration viewing and editing
- Model synchronization controls

For detailed model manager documentation, see [`MODEL_MANAGER.md`](MODEL_MANAGER.md).

## 🧪 Testing

### Manual Testing

1. **Start the server:**
   ```bash
   make run
   ```

2. **Test health endpoint:**
   ```bash
   curl http://localhost:8080/health
   ```

3. **Test chat endpoint:**
   ```bash
   curl -X POST http://localhost:8080/v1/chat \
     -H "Content-Type: application/json" \
     -d '{"message":"Hello","session_id":"test-session","model":"llama2:7b","stream":false}'
   ```

4. **Test model management:**
   ```bash
   # Sync models with Ollama
   curl -X POST http://localhost:8080/v1/models/sync
   
   # List available models
   curl http://localhost:8080/v1/models?available=true
   
   # Get model details
   curl http://localhost:8080/v1/models/{model-id}
   ```

5. **Test web interface:**
   - Open http://localhost:8080 in your browser
   - Use the chat interface to send messages
   - Click the "🤖 Models" tab to manage models
   - Sync models and configure settings

### Docker Testing

```bash
make docker-run
curl http://localhost:8080/health
```

## 📝 Notes

- **Model Management**: Full implementation with Ollama integration, configuration management, and web UI
- **Chat Functionality**: Complete streaming and non-streaming chat with message persistence
- **Authentication**: Not implemented - planned for future releases
- **Rate Limiting**: Basic structure in place, implementation pending
- **Metrics**: Logging foundation ready for metrics integration

## 🔧 Model Manager

The application includes a comprehensive model management system:

- **Automatic Discovery**: Syncs with Ollama to discover available models
- **Configuration**: Per-model settings for temperature, context length, etc.
- **Usage Tracking**: Statistics on model usage and performance
- **Web Interface**: User-friendly model management in the browser
- **API Integration**: RESTful API for programmatic model management

See [`MODEL_MANAGER.md`](MODEL_MANAGER.md) for complete documentation.

## 🤝 Contributing

1. Follow the architectural patterns established in `ARCHITECTURE.md`
2. Use the provided Makefile for development tasks
3. Ensure all tests pass before submitting changes
4. Follow Go best practices and formatting standards

## 📄 License

[Add your license information here]
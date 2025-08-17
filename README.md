# OllamaPilot - Semantic Memory Chat Application

A modern chat application with semantic memory capabilities, built with Go, PostgreSQL, and pgvector for intelligent conversation management.

## 🧠 Features

- **Semantic Memory**: Vector-based similarity search across conversations
- **Memory Gap Detection**: Automatic detection of context discontinuities
- **Conversation Summarization**: Intelligent memory consolidation
- **Real-time Chat**: Streaming and non-streaming chat with Ollama
- **Model Management**: Comprehensive LLM model configuration
- **Docker Ready**: Complete containerized deployment

## 🏗️ Architecture

- **Backend**: Go with Chi router and structured logging
- **Database**: PostgreSQL with pgvector extension for vector operations
- **LLM Integration**: Ollama for chat and embedding generation
- **Frontend**: Modern web interface with real-time updates
- **Deployment**: Docker Compose with health checks

## 📁 Project Structure

```
ollamapilot/
├── cmd/api/main.go                 # Application entry point
├── internal/
│   ├── api/                        # HTTP handlers and routing
│   ├── config/                     # Configuration management
│   ├── database/                   # Database connections (PostgreSQL)
│   ├── models/                     # Data models
│   ├── services/                   # Business logic services
│   │   ├── chat.go                 # Chat orchestration
│   │   ├── embedding.go            # Embedding generation
│   │   └── semantic_memory.go      # Semantic memory operations
│   └── utils/                      # Utilities and helpers
├── migrations/postgres/            # PostgreSQL migrations with pgvector
├── web/                           # Frontend assets
├── docker-compose.yml             # Complete Docker setup
├── .env                          # Environment configuration
└── .env.example                  # Configuration template
```

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Ollama (for local development)

### 1. Clone and Setup

```bash
git clone <repository>
cd ollamapilot
cp .env.example .env
# Edit .env if needed
```

### 2. Start Services

```bash
docker-compose up -d
```

This starts:
- PostgreSQL with pgvector extension
- OllamaPilot API server
- Ollama LLM runtime

### 3. Pull Models

```bash
# Pull chat model
docker exec -it ollamapilot-ollama ollama pull llama3.2:3b

# Pull embedding model for semantic memory
docker exec -it ollamapilot-ollama ollama pull nomic-embed-text
```

### 4. Access Application

- **Web UI**: http://localhost:8080
- **API**: http://localhost:8080/v1/
- **Health Check**: http://localhost:8080/health

## 🔍 API Endpoints

### Chat API
- `POST /v1/chat` - Send chat message (streaming/non-streaming)
- `GET /v1/sessions` - List chat sessions
- `GET /v1/sessions/{id}/messages` - Get session messages
- `DELETE /v1/sessions/{id}` - Delete session

### Semantic Memory API
- `POST /v1/memory/search` - Semantic search across conversations
- `GET /v1/memory/summaries` - Get conversation summaries
- `POST /v1/memory/summaries` - Create memory summary
- `GET /v1/memory/gaps/{sessionID}` - Detect memory gaps

### Model Management API
- `GET /v1/models` - List available models
- `POST /v1/models/sync` - Sync with Ollama
- `PUT /v1/models/{id}` - Update model settings

### Health Checks
- `GET /health` - Comprehensive health check
- `GET /ready` - Readiness probe
- `GET /live` - Liveness probe

## 🧠 Semantic Memory Features

### Vector Search
```bash
curl -X POST http://localhost:8080/v1/memory/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "machine learning algorithms",
    "limit": 5
  }'
```

### Memory Summaries
```bash
curl -X POST http://localhost:8080/v1/memory/summaries \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "abc123",
    "summary_type": "conversation",
    "content": "Discussion about ML algorithms and applications"
  }'
```

### Memory Gap Detection
```bash
curl http://localhost:8080/v1/memory/gaps/abc123?threshold=1h
```

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_TYPE` | `postgres` | Database type (postgres) |
| `DB_HOST` | `postgres` | PostgreSQL host |
| `DB_NAME` | `ollamapilot` | Database name |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `OLLAMA_HOST` | `ollama:11434` | Ollama service host |
| `PORT` | `8080` | Server port |
| `LOG_LEVEL` | `info` | Logging level |

### Development Setup

For local development without Docker:

```bash
# Copy environment file
cp .env.example .env

# Edit for local development
# Change DB_HOST=localhost, OLLAMA_HOST=localhost:11434

# Start PostgreSQL with pgvector
docker run -d --name postgres-pgvector \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  pgvector/pgvector:pg16

# Start Ollama
ollama serve

# Run application
go run cmd/api/main.go
```

## 🗄️ Database Schema

### Core Tables
- **sessions**: Chat session metadata
- **messages**: Individual chat messages
- **models**: LLM model configurations

### Semantic Memory Tables
- **message_embeddings**: Vector embeddings for semantic search
- **memory_summaries**: Conversation summaries and consolidation
- **memory_gaps**: Context discontinuity tracking
- **semantic_topics**: Topic categorization

### Vector Operations
The application uses pgvector for:
- Cosine similarity search
- IVFFlat indexes for performance
- 1536-dimensional embeddings (configurable)

## 🐳 Docker Services

### PostgreSQL
- Image: `pgvector/pgvector:pg16`
- Extensions: pgvector for vector operations
- Persistent storage with health checks

### API Server
- Built from source with multi-stage Dockerfile
- Health checks and graceful shutdown
- Environment-based configuration

### Ollama
- Official Ollama image with GPU support
- Model persistence and parallel processing
- Configurable resource limits

## 📊 Monitoring

### Health Endpoints
```bash
# Full health check
curl http://localhost:8080/health

# Database connectivity
curl http://localhost:8080/ready

# Application liveness
curl http://localhost:8080/live
```

### Logs
```bash
# View all logs
docker-compose logs -f

# Specific service logs
docker-compose logs -f api
docker-compose logs -f postgres
docker-compose logs -f ollama
```

## 🔧 Development

### Available Commands
```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Rebuild and restart
docker-compose up --build -d

# Database shell
docker exec -it ollamapilot-postgres psql -U postgres -d ollamapilot
```

### Testing
```bash
# Test health
curl http://localhost:8080/health

# Test chat
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"Hello","session_id":"test","model":"llama3.2:3b","stream":false}'

# Test semantic search
curl -X POST http://localhost:8080/v1/memory/search \
  -H "Content-Type: application/json" \
  -d '{"query":"hello","limit":5}'
```

## 📚 Documentation

- **SEMANTIC_MEMORY.md**: Detailed semantic memory implementation
- **DOCKER_SETUP.md**: Docker deployment and troubleshooting
- **ARCHITECTURE.md**: System architecture overview
- **MODEL_MANAGER.md**: Model management documentation

## 🚧 Features

✅ **Implemented:**
- Semantic memory with vector search
- Memory gap detection and bridging
- Conversation summarization
- Real-time chat with streaming
- Model management and configuration
- Docker containerization
- PostgreSQL with pgvector
- Health monitoring and logging

🔄 **Planned:**
- Cross-session learning
- Hierarchical memory structures
- Advanced topic clustering
- Memory compression algorithms
- Enhanced web interface

## 🤝 Contributing

1. Follow Go best practices and formatting
2. Use the provided Docker setup for development
3. Ensure all health checks pass
4. Update documentation for new features

## 📄 License

[Add your license information here]
# Semantic Memory Implementation

This document describes the semantic memory and vector search capabilities added to OllamaPilot.

## Overview

The semantic memory system enhances the chat application with:
- **Vector embeddings** for semantic search
- **Memory consolidation** through conversation summarization
- **Memory gap detection** for context continuity
- **Intelligent context retrieval** for improved responses

## Architecture

### Database Migration

The application now supports both SQLite and PostgreSQL:
- **SQLite**: Legacy support for existing installations
- **PostgreSQL + pgvector**: Recommended for semantic memory features

### Key Components

1. **EmbeddingService** (`internal/services/embedding.go`)
   - Generates text embeddings using Ollama models
   - Supports models like `nomic-embed-text`, `all-minilm`, `mxbai-embed-large`
   - Handles batch embedding generation

2. **SemanticMemoryService** (`internal/services/semantic_memory.go`)
   - Stores and retrieves message embeddings
   - Performs vector similarity search
   - Creates conversation summaries
   - Detects memory gaps

3. **Database Layer**
   - PostgreSQL with pgvector extension for vector operations
   - Optimized indexes for similarity search
   - Migration system for schema updates

## Database Schema

### New Tables

#### `message_embeddings`
Stores vector embeddings for messages:
```sql
CREATE TABLE message_embeddings (
    id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL,
    embedding vector(1536),
    model_used TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

#### `memory_summaries`
Stores conversation summaries:
```sql
CREATE TABLE memory_summaries (
    id TEXT PRIMARY KEY,
    session_id TEXT,
    summary_type TEXT NOT NULL,
    title TEXT,
    content TEXT NOT NULL,
    embedding vector(1536),
    relevance_score REAL DEFAULT 0.0,
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    message_count INTEGER DEFAULT 0
);
```

#### `memory_gaps`
Tracks context discontinuities:
```sql
CREATE TABLE memory_gaps (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    gap_start TIMESTAMP WITH TIME ZONE NOT NULL,
    gap_end TIMESTAMP WITH TIME ZONE NOT NULL,
    context_summary TEXT,
    bridge_content TEXT,
    gap_type TEXT DEFAULT 'temporal'
);
```

#### `semantic_topics`
Categorizes conversations by topic:
```sql
CREATE TABLE semantic_topics (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    embedding vector(1536),
    message_count INTEGER DEFAULT 0
);
```

## API Endpoints

### Semantic Search
```http
POST /v1/memory/search
Content-Type: application/json

{
  "query": "What did we discuss about machine learning?",
  "session_id": "optional-session-filter",
  "limit": 10
}
```

### Memory Summaries
```http
GET /v1/memory/summaries?session_id=abc&type=conversation
POST /v1/memory/summaries
{
  "session_id": "abc",
  "summary_type": "conversation",
  "title": "ML Discussion",
  "content": "Summary of machine learning conversation...",
  "message_count": 15
}
```

### Memory Gaps
```http
GET /v1/memory/gaps/{sessionID}?threshold=1h
```

## Configuration

### Environment Variables

For PostgreSQL with semantic memory:
```bash
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ollamapilot
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSL_MODE=disable
```

For SQLite (legacy):
```bash
DB_TYPE=sqlite
DB_PATH=./data/chat.db
```

## Deployment

### Using Docker Compose with PostgreSQL

1. **Start with PostgreSQL:**
   ```bash
   docker-compose -f docker-compose.postgres.yml up -d
   ```

2. **Environment file:**
   ```bash
   cp .env.postgres .env
   ```

3. **Verify pgvector extension:**
   ```sql
   SELECT * FROM pg_extension WHERE extname = 'vector';
   ```

### Manual Setup

1. **Install PostgreSQL with pgvector:**
   ```bash
   # Ubuntu/Debian
   sudo apt install postgresql-16-pgvector
   
   # Or use Docker
   docker run -d --name postgres-pgvector \
     -e POSTGRES_PASSWORD=postgres \
     -p 5432:5432 \
     pgvector/pgvector:pg16
   ```

2. **Create database:**
   ```sql
   CREATE DATABASE ollamapilot;
   CREATE EXTENSION vector;
   ```

3. **Run migrations:**
   ```bash
   go run cmd/api/main.go
   ```

## Usage Examples

### Semantic Search
The system automatically generates embeddings for all messages and enables semantic search:

```javascript
// Search for similar messages
const response = await fetch('/v1/memory/search', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    query: 'machine learning algorithms',
    limit: 5
  })
});

const results = await response.json();
// Returns messages semantically similar to the query
```

### Enhanced Chat Context
The chat service automatically retrieves relevant context:

1. User sends a message
2. System generates embedding for the message
3. Searches for semantically similar past messages
4. Includes relevant context in the LLM prompt
5. Stores new message embedding for future searches

### Memory Gap Detection
Automatically detects and handles conversation gaps:

```javascript
// Get memory gaps for a session
const gaps = await fetch('/v1/memory/gaps/session-123?threshold=2h');
// Returns detected gaps longer than 2 hours
```

## Performance Considerations

### Vector Indexes
The system uses IVFFlat indexes for efficient similarity search:
```sql
CREATE INDEX idx_message_embeddings_vector 
ON message_embeddings USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);
```

### Embedding Models
- **nomic-embed-text**: 768 dimensions, good general purpose
- **all-minilm**: 384 dimensions, faster but less accurate
- **mxbai-embed-large**: 1024 dimensions, high accuracy

### Scaling
- Use connection pooling for high concurrency
- Consider read replicas for search-heavy workloads
- Monitor vector index performance and adjust `lists` parameter

## Troubleshooting

### Common Issues

1. **pgvector not installed:**
   ```
   ERROR: extension "vector" is not available
   ```
   Solution: Install pgvector extension

2. **Embedding model not available:**
   ```
   ERROR: embedding model nomic-embed-text is not available
   ```
   Solution: Pull the model with Ollama: `ollama pull nomic-embed-text`

3. **Vector dimension mismatch:**
   ```
   ERROR: vector dimension mismatch
   ```
   Solution: Ensure consistent embedding model usage

### Monitoring

Monitor these metrics:
- Embedding generation latency
- Vector search performance
- Memory gap detection frequency
- Storage usage growth

## Future Enhancements

Planned improvements:
- **Hierarchical memory**: Multi-level summarization
- **Adaptive forgetting**: Automatic cleanup of old memories
- **Cross-session learning**: Global knowledge base
- **Real-time clustering**: Dynamic topic discovery
- **Memory compression**: Efficient long-term storage
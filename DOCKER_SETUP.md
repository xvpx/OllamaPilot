# Docker Setup for OllamaPilot with Semantic Memory

This guide covers running OllamaPilot with PostgreSQL and pgvector in Docker.

## Quick Start

```bash
# Start all services (PostgreSQL + pgvector, API, Ollama)
docker-compose up -d
```

## PostgreSQL + pgvector Setup

### Prerequisites
- Docker and Docker Compose installed
- Ollama running on host machine (accessible via `host.docker.internal:11434`)

### Configuration Files

#### Environment Variables (`.env.example`)
```bash
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ollamapilot
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSL_MODE=disable
OLLAMA_HOST=localhost:11434
PORT=8080
ENV=development
LOG_LEVEL=info
```

#### Docker Compose (`docker-compose.yml`)
- PostgreSQL 16 with pgvector extension
- Automatic database initialization
- Health checks for proper startup order
- Persistent data volumes

### Startup Process

1. **Start services:**
   ```bash
   docker-compose up -d
   ```

2. **Check logs:**
   ```bash
   # PostgreSQL logs
   docker-compose logs postgres
   
   # Application logs
   docker-compose logs api
   
   # Ollama logs
   docker-compose logs ollama
   ```

3. **Verify pgvector extension:**
   ```bash
   docker exec -it ollamapilot-postgres psql -U postgres -d ollamapilot -c "SELECT * FROM pg_extension WHERE extname = 'vector';"
   ```

4. **Pull embedding models:**
   ```bash
   # Pull required embedding models
   docker exec -it ollamapilot-ollama ollama pull nomic-embed-text
   docker exec -it ollamapilot-ollama ollama pull llama3.2:3b
   ```

### Database Migrations

Migrations run automatically on startup:
- `001_initial_schema.sql` - Creates tables with pgvector support
- `002_add_indexes.sql` - Adds performance indexes
- `003_add_model_management.sql` - Model management tables

### Accessing the Application

- **Web UI:** http://localhost:8080
- **API:** http://localhost:8080/v1/
- **Health Check:** http://localhost:8080/health

## Semantic Memory Features

### API Endpoints

#### Semantic Search
```bash
curl -X POST http://localhost:8080/v1/memory/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "machine learning algorithms",
    "limit": 5
  }'
```

#### Memory Summaries
```bash
# Get summaries
curl http://localhost:8080/v1/memory/summaries?session_id=abc123

# Create summary
curl -X POST http://localhost:8080/v1/memory/summaries \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "abc123",
    "summary_type": "conversation",
    "content": "Discussion about ML algorithms and their applications"
  }'
```

#### Memory Gaps
```bash
curl http://localhost:8080/v1/memory/gaps/abc123?threshold=1h
```

### Embedding Models

Ensure these models are available in Ollama:
```bash
# Pull embedding models
ollama pull nomic-embed-text
ollama pull all-minilm
ollama pull mxbai-embed-large
```

## Troubleshooting

### Common Issues

1. **PostgreSQL connection failed:**
   ```bash
   # Check if PostgreSQL is ready
   docker-compose -f docker-compose.postgres.yml ps
   
   # Check PostgreSQL logs
   docker-compose -f docker-compose.postgres.yml logs postgres
   ```

2. **pgvector extension missing:**
   ```bash
   # Verify extension is installed
   docker exec -it ollamapilot-postgres psql -U postgres -c "CREATE EXTENSION IF NOT EXISTS vector;"
   ```

3. **Ollama not accessible:**
   ```bash
   # Test Ollama connectivity from container
   docker exec -it ollamapilot-app curl http://host.docker.internal:11434/api/tags
   ```

4. **Migration failures:**
   ```bash
   # Check application logs for migration errors
   docker-compose -f docker-compose.postgres.yml logs app | grep migration
   ```

### Health Checks

```bash
# Application health
curl http://localhost:8080/health

# Database health
curl http://localhost:8080/ready

# Liveness probe
curl http://localhost:8080/live
```

### Data Persistence

Data is persisted in Docker volumes:
- `postgres_data` - PostgreSQL data
- Application logs are ephemeral (use external logging for production)

### Backup and Restore

```bash
# Backup PostgreSQL data
docker exec ollamapilot-postgres pg_dump -U postgres ollamapilot > backup.sql

# Restore PostgreSQL data
docker exec -i ollamapilot-postgres psql -U postgres ollamapilot < backup.sql
```

## Development

### Local Development with Docker

1. **Build and run:**
   ```bash
   docker-compose -f docker-compose.postgres.yml up --build
   ```

2. **Watch logs:**
   ```bash
   docker-compose -f docker-compose.postgres.yml logs -f app
   ```

3. **Access database:**
   ```bash
   docker exec -it ollamapilot-postgres psql -U postgres -d ollamapilot
   ```

### Environment Switching

Switch between SQLite and PostgreSQL:
```bash
# PostgreSQL
cp .env.postgres .env
docker-compose -f docker-compose.postgres.yml up -d

# SQLite
cp .env.example .env
docker-compose up -d
```

## Production Deployment

### Security Considerations

1. **Change default passwords:**
   ```bash
   # Update .env.postgres
   DB_PASSWORD=your-secure-password
   ```

2. **Enable SSL:**
   ```bash
   DB_SSL_MODE=require
   ```

3. **Use secrets management:**
   - Docker secrets
   - Kubernetes secrets
   - External secret managers

### Scaling

1. **Database scaling:**
   - Read replicas for search workloads
   - Connection pooling (PgBouncer)
   - Monitoring and alerting

2. **Application scaling:**
   ```bash
   docker-compose -f docker-compose.postgres.yml up -d --scale app=3
   ```

### Monitoring

Monitor these metrics:
- PostgreSQL performance
- Vector search latency
- Embedding generation time
- Memory usage
- Storage growth

## Migration from SQLite

To migrate existing SQLite data to PostgreSQL:

1. **Export SQLite data:**
   ```bash
   # This would require a custom migration script
   # Not implemented in this version
   ```

2. **Import to PostgreSQL:**
   ```bash
   # Custom import process needed
   # Consider data transformation for vector embeddings
   ```

Note: Direct migration tools are not included in this implementation. Consider this for future versions.
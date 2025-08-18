# Chat Ollama MVP - Docker Deployment Guide

## Overview

This guide provides comprehensive instructions for deploying the Chat Ollama MVP using Docker containers. The setup includes a production-ready Go API server and Ollama service with proper networking, health checks, and data persistence.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   API Service   │    │  Ollama Service │
│   (Port 8081)   │◄──►│  (Port 11435)   │
│                 │    │                 │
│ - Go API Server │    │ - LLM Runtime   │
│ - PostgreSQL DB │    │ - Model Storage │
│ - Health Checks │    │ - Health Checks │
└─────────────────┘    └─────────────────┘
         │                       │
         └───────────────────────┘
                     │
            ┌─────────────────┐
            │   Docker Volumes │
            │                 │
            │ - chat_data     │
            │ - ollama_models │
            └─────────────────┘
```

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- At least 4GB RAM available
- 10GB free disk space for models

## Quick Start

### 1. Clone and Navigate
```bash
cd /path/to/chat_ollama
```

### 2. Build and Start Services
```bash
# Using docker-compose
docker-compose up --build -d

# Or using the deployment script
chmod +x docker/deploy.sh
./docker/deploy.sh
```

### 3. Verify Deployment
```bash
# Check service status
docker-compose ps

# Test API health
curl http://localhost:8081/health

# Test Ollama health
curl http://localhost:11435/api/tags
```

## Deployment Scripts

### Production Deployment Script
Use the provided deployment script for automated setup:

```bash
cd docker
./deploy.sh
```

### Development Script
For development with additional utilities:

```bash
cd docker
./dev.sh up
```

Available dev commands:
- `./dev.sh up` - Start development environment
- `./dev.sh down` - Stop services
- `./dev.sh logs` - Show service logs
- `./dev.sh build` - Rebuild containers
- `./dev.sh clean` - Clean up containers and volumes
- `./dev.sh status` - Show service status
- `./dev.sh shell` - Open shell in API container
- `./dev.sh test` - Run health checks

## Configuration

### Environment Variables

The application supports the following environment variables:

#### Server Configuration
- `PORT=8080` - API server port (internal)
- `HOST=0.0.0.0` - API server host
- `ENV=production` - Environment mode

#### Database Configuration
- PostgreSQL database configuration via environment variables

#### Ollama Configuration
- `OLLAMA_HOST=ollama:11434` - Ollama service endpoint
- `OLLAMA_TIMEOUT=30s` - Request timeout
- `OLLAMA_NUM_PARALLEL=4` - Parallel processing
- `OLLAMA_MAX_LOADED_MODELS=1` - Memory management

#### Logging Configuration
- `LOG_LEVEL=info` - Logging level
- `LOG_FORMAT=json` - Log format

### Port Configuration

Default ports (can be modified in docker-compose.yml):
- **API Service**: `8081:8080` (host:container)
- **Ollama Service**: `11435:11434` (host:container)

## Data Persistence

### Volumes

Two Docker volumes ensure data persistence:

1. **postgres_data**: PostgreSQL database storage
   - Path: `/var/lib/postgresql/data` in PostgreSQL container
   - Contains: `chat.db` and related files

2. **ollama_models**: Model storage
   - Path: `/root/.ollama` in Ollama container
   - Contains: Downloaded LLM models

### Backup and Restore

#### Backup
```bash
# Backup database
docker cp chat_ollama-api-1:/data/chat.db ./backup/chat.db

# Backup models (optional, models can be re-downloaded)
docker cp chat_ollama-ollama-1:/root/.ollama ./backup/ollama_models
```

#### Restore
```bash
# Restore database
docker cp ./backup/chat.db chat_ollama-api-1:/data/chat.db

# Restart API service to reload
docker-compose restart api
```

## Health Monitoring

### Health Check Endpoints

#### API Service Health
```bash
curl http://localhost:8081/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2025-08-16T12:40:06Z",
  "services": {
    "database": "healthy",
    "ollama": "healthy"
  },
  "metadata": {
    "postgres_version": "PostgreSQL 16.x"
  }
}
```

#### Ollama Service Health
```bash
curl http://localhost:11435/api/tags
```

### Docker Health Checks

Both services include built-in Docker health checks:
- **Interval**: 30 seconds
- **Timeout**: 10 seconds
- **Retries**: 3
- **Start Period**: 10s (API), 30s (Ollama)

## Security Features

### Container Security
- **Non-root user**: API runs as user `appuser` (UID 1001)
- **Read-only filesystem**: API container filesystem is read-only except for `/data`
- **No new privileges**: Security option prevents privilege escalation
- **Minimal base image**: Alpine Linux for reduced attack surface

### Network Security
- **Custom network**: Isolated Docker network `chat_ollama_network`
- **Internal communication**: Services communicate via container names
- **Port isolation**: Only necessary ports exposed to host

### Environment Security
- **No secrets in images**: All sensitive data via environment variables
- **Secure defaults**: Production-ready configuration out of the box

## Troubleshooting

### Common Issues

#### Port Conflicts
If ports 8081 or 11435 are in use:
```bash
# Check what's using the ports
sudo lsof -i :8081
sudo lsof -i :11435

# Modify ports in docker-compose.yml
# Change "8081:8080" to "8082:8080" for API
# Change "11435:11434" to "11436:11434" for Ollama
```

#### Service Won't Start
```bash
# Check logs
docker-compose logs api
docker-compose logs ollama

# Check service status
docker-compose ps

# Restart services
docker-compose restart
```

#### Database Issues
```bash
# Check database file permissions
docker exec chat_ollama-api-1 ls -la /data/

# Reset database (WARNING: deletes all data)
docker-compose down
docker volume rm chat_ollama_chat_data
docker-compose up -d
```

#### Ollama Model Issues
```bash
# Check available models
curl http://localhost:11435/api/tags

# Pull a model manually
docker exec chat_ollama-ollama-1 ollama pull llama2:7b

# Check Ollama logs
docker-compose logs ollama
```

### Performance Tuning

#### Memory Optimization
```yaml
# In docker-compose.yml, add memory limits
services:
  api:
    mem_limit: 512m
  ollama:
    mem_limit: 4g  # Adjust based on model size
```

#### CPU Optimization
```yaml
# Limit CPU usage
services:
  api:
    cpus: '0.5'
  ollama:
    cpus: '2.0'
```

## Scaling and Production

### Production Considerations

1. **Resource Allocation**
   - Minimum 4GB RAM for Ollama
   - Additional 2GB per concurrent model
   - SSD storage recommended for performance

2. **Load Balancing**
   - Use nginx or traefik for load balancing
   - Scale API service horizontally
   - Keep Ollama service single instance per model

3. **Monitoring**
   - Implement log aggregation (ELK stack)
   - Use Prometheus for metrics
   - Set up alerting for health check failures

4. **Backup Strategy**
   - Regular database backups
   - Model storage backup (optional)
   - Configuration backup

### Example Production docker-compose.yml
```yaml
version: '3.8'
services:
  api:
    image: chat_ollama-api:latest
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 512M
        reservations:
          memory: 256M
    # ... rest of configuration
```

## API Usage Examples

### Basic Health Check
```bash
curl -X GET http://localhost:8081/health
```

### Chat Request (when implemented)
```bash
curl -X POST http://localhost:8081/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Hello, how are you?",
    "model": "llama2:7b",
    "stream": false
  }'
```

## Maintenance

### Regular Maintenance Tasks

1. **Update Images**
   ```bash
   docker-compose pull
   docker-compose up -d
   ```

2. **Clean Up**
   ```bash
   # Remove unused images
   docker image prune
   
   # Remove unused volumes (careful!)
   docker volume prune
   ```

3. **Log Rotation**
   ```bash
   # Configure log rotation in docker-compose.yml
   logging:
     driver: "json-file"
     options:
       max-size: "10m"
       max-file: "3"
   ```

## Support

For issues and questions:
1. Check the troubleshooting section above
2. Review Docker and application logs
3. Verify system requirements
4. Check network connectivity between containers

## Version Information

- **Docker Image**: Built from source
- **Base Images**: 
  - API: `golang:1.22-alpine` → `alpine:latest`
  - Ollama: `ollama/ollama:latest`
- **Architecture**: Multi-stage Docker build
- **Database**: PostgreSQL 16.x with pgvector
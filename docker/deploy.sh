#!/bin/bash

# Chat Ollama Docker Deployment Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.yml"
ENV_FILE="../.env.production"
DATA_DIR="./data"

echo -e "${GREEN}🚀 Chat Ollama Docker Deployment${NC}"
echo "=================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose > /dev/null 2>&1; then
    echo -e "${RED}❌ docker-compose is not installed. Please install it and try again.${NC}"
    exit 1
fi

# Create data directory if it doesn't exist
if [ ! -d "$DATA_DIR" ]; then
    echo -e "${YELLOW}📁 Creating data directory...${NC}"
    mkdir -p "$DATA_DIR"
    chmod 755 "$DATA_DIR"
fi

# Copy environment file if it doesn't exist
if [ ! -f ".env" ] && [ -f "$ENV_FILE" ]; then
    echo -e "${YELLOW}📋 Copying production environment file...${NC}"
    cp "$ENV_FILE" ".env"
fi

# Build and start services
echo -e "${YELLOW}🔨 Building and starting services...${NC}"
docker-compose -f "$COMPOSE_FILE" up --build -d

# Wait for services to be healthy
echo -e "${YELLOW}⏳ Waiting for services to be healthy...${NC}"
timeout=120
counter=0

while [ $counter -lt $timeout ]; do
    if docker-compose -f "$COMPOSE_FILE" ps | grep -q "healthy"; then
        echo -e "${GREEN}✅ Services are healthy!${NC}"
        break
    fi
    
    if [ $counter -eq $timeout ]; then
        echo -e "${RED}❌ Timeout waiting for services to be healthy${NC}"
        docker-compose -f "$COMPOSE_FILE" logs
        exit 1
    fi
    
    echo -n "."
    sleep 2
    counter=$((counter + 2))
done

# Show service status
echo -e "\n${GREEN}📊 Service Status:${NC}"
docker-compose -f "$COMPOSE_FILE" ps

# Show logs
echo -e "\n${GREEN}📝 Recent Logs:${NC}"
docker-compose -f "$COMPOSE_FILE" logs --tail=10

echo -e "\n${GREEN}🎉 Deployment completed successfully!${NC}"
echo -e "API is available at: ${YELLOW}http://localhost:8080${NC}"
echo -e "Ollama is available at: ${YELLOW}http://localhost:11434${NC}"
echo -e "Health check: ${YELLOW}http://localhost:8080/health${NC}"
echo ""
echo -e "To view logs: ${YELLOW}docker-compose -f $COMPOSE_FILE logs -f${NC}"
echo -e "To stop services: ${YELLOW}docker-compose -f $COMPOSE_FILE down${NC}"
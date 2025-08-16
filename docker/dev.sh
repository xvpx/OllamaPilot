#!/bin/bash

# Chat Ollama Development Docker Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="../docker-compose.yml"
DATA_DIR="../data"

echo -e "${BLUE}🛠️  Chat Ollama Development Environment${NC}"
echo "======================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi

# Create data directory if it doesn't exist
if [ ! -d "$DATA_DIR" ]; then
    echo -e "${YELLOW}📁 Creating data directory...${NC}"
    mkdir -p "$DATA_DIR"
    chmod 755 "$DATA_DIR"
fi

# Function to show usage
show_usage() {
    echo -e "${GREEN}Usage: $0 [COMMAND]${NC}"
    echo ""
    echo -e "${YELLOW}Commands:${NC}"
    echo "  up      - Start development environment"
    echo "  down    - Stop development environment"
    echo "  logs    - Show service logs"
    echo "  build   - Rebuild containers"
    echo "  clean   - Clean up containers and volumes"
    echo "  status  - Show service status"
    echo "  shell   - Open shell in API container"
    echo "  test    - Run health checks"
    echo ""
}

# Function to start services
start_services() {
    echo -e "${YELLOW}🚀 Starting development environment...${NC}"
    cd ..
    docker-compose up --build -d
    
    echo -e "${YELLOW}⏳ Waiting for services to be ready...${NC}"
    sleep 10
    
    echo -e "${GREEN}📊 Service Status:${NC}"
    docker-compose ps
    
    echo -e "\n${GREEN}🎉 Development environment is ready!${NC}"
    echo -e "API: ${YELLOW}http://localhost:8080${NC}"
    echo -e "Ollama: ${YELLOW}http://localhost:11434${NC}"
    echo -e "Health: ${YELLOW}http://localhost:8080/health${NC}"
}

# Function to stop services
stop_services() {
    echo -e "${YELLOW}🛑 Stopping development environment...${NC}"
    cd ..
    docker-compose down
    echo -e "${GREEN}✅ Services stopped${NC}"
}

# Function to show logs
show_logs() {
    echo -e "${YELLOW}📝 Showing service logs...${NC}"
    cd ..
    docker-compose logs -f
}

# Function to rebuild containers
rebuild_containers() {
    echo -e "${YELLOW}🔨 Rebuilding containers...${NC}"
    cd ..
    docker-compose down
    docker-compose build --no-cache
    docker-compose up -d
    echo -e "${GREEN}✅ Containers rebuilt${NC}"
}

# Function to clean up
clean_up() {
    echo -e "${YELLOW}🧹 Cleaning up containers and volumes...${NC}"
    cd ..
    docker-compose down -v
    docker system prune -f
    echo -e "${GREEN}✅ Cleanup completed${NC}"
}

# Function to show status
show_status() {
    echo -e "${YELLOW}📊 Service Status:${NC}"
    cd ..
    docker-compose ps
    echo ""
    echo -e "${YELLOW}🔍 Health Checks:${NC}"
    curl -s http://localhost:8080/health | jq . 2>/dev/null || echo "API not responding"
    curl -s http://localhost:11434/api/tags | jq . 2>/dev/null || echo "Ollama not responding"
}

# Function to open shell
open_shell() {
    echo -e "${YELLOW}🐚 Opening shell in API container...${NC}"
    cd ..
    docker-compose exec api sh
}

# Function to run tests
run_tests() {
    echo -e "${YELLOW}🧪 Running health checks...${NC}"
    
    echo -n "Testing API health endpoint... "
    if curl -s -f http://localhost:8080/health > /dev/null; then
        echo -e "${GREEN}✅ PASS${NC}"
    else
        echo -e "${RED}❌ FAIL${NC}"
    fi
    
    echo -n "Testing Ollama endpoint... "
    if curl -s -f http://localhost:11434/api/tags > /dev/null; then
        echo -e "${GREEN}✅ PASS${NC}"
    else
        echo -e "${RED}❌ FAIL${NC}"
    fi
}

# Main script logic
case "${1:-}" in
    "up")
        start_services
        ;;
    "down")
        stop_services
        ;;
    "logs")
        show_logs
        ;;
    "build")
        rebuild_containers
        ;;
    "clean")
        clean_up
        ;;
    "status")
        show_status
        ;;
    "shell")
        open_shell
        ;;
    "test")
        run_tests
        ;;
    *)
        show_usage
        ;;
esac
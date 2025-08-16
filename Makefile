.PHONY: build run test clean docker-build docker-run dev

# Build the application
build:
	CGO_ENABLED=1 go build -o bin/api ./cmd/api

# Run the application locally
run: build
	./bin/api

# Run in development mode with hot reload (requires air)
dev:
	air

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf data/

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build Docker image
docker-build:
	docker build -t chat-ollama-api .

# Run with Docker Compose
docker-run:
	docker-compose up --build

# Run Docker Compose in background
docker-up:
	docker-compose up -d --build

# Stop Docker Compose
docker-down:
	docker-compose down

# View Docker logs
docker-logs:
	docker-compose logs -f api

# Create data directory
setup:
	mkdir -p data

# Run database migrations manually
migrate: build setup
	./bin/api

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Security scan (requires gosec)
security:
	gosec ./...

# Generate API documentation
docs:
	swag init -g cmd/api/main.go

# Install development tools
install-tools:
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# Help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  run          - Build and run the application"
	@echo "  dev          - Run in development mode with hot reload"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  docker-up    - Run Docker Compose in background"
	@echo "  docker-down  - Stop Docker Compose"
	@echo "  docker-logs  - View Docker logs"
	@echo "  setup        - Create data directory"
	@echo "  migrate      - Run database migrations"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  security     - Run security scan"
	@echo "  docs         - Generate API documentation"
	@echo "  install-tools- Install development tools"
	@echo "  help         - Show this help"
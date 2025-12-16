.PHONY: help build run test clean docker-build docker-run

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building lucidRAG..."
	@go build -o bin/lucidrag cmd/api/main.go

run: ## Run the application
	@echo "Running lucidRAG..."
	@go run cmd/api/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	@echo "Running linter..."
	@go vet ./...
	@go fmt ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t lucidrag:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker-compose up -d

docker-stop: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose down

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

install: ## Install the application
	@echo "Installing lucidRAG..."
	@go install ./cmd/api

.DEFAULT_GOAL := help

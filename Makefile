.PHONY: help setup deps db-up db-down server scheduler test clean

help: ## Show available commands
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Install Go dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

setup: deps ## Setup project (install deps + copy env)
	@echo "Setting up project..."
	cp .env.example .env
	mkdir -p logs

db-up: ## Start database and redis only
	@echo "Starting database services..."
	docker compose up -d postgres redis
	@sleep 3
	@echo "✅ Database ready at localhost:5432"
	@echo "✅ Redis ready at localhost:6379"

db-down: ## Stop database services
	docker compose down

server: ## Run server locally with air (hot reload)
	@echo "Starting server with hot reload..."
	air -c .air.toml

scheduler: ## Run scheduler locally
	@echo "Starting scheduler..."
	go run cmd/scheduler/main.go

dev: db-up ## Start everything for development
	@echo "💡 Run 'make server' in another terminal"
	@echo "💡 Run 'make scheduler' in another terminal if needed"

test: ## Run tests
	go test -v ./...

test-race: ## Run tests with race detection
	go test -race -v ./...

clean: ## Clean up
	docker compose down -v
	go clean
	rm -rf logs/*.log

demo: ## Run demo (assumes server is running)
	@curl -s http://localhost:8080/health || echo "Start server first with: make server"
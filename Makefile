.PHONY: help setup deps db-up db-down server scheduler test clean dev logs

help: ## Show available commands
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Setup project (copy env + create logs dir)
	@echo "Setting up project..."
	cp .env.example .env
	mkdir -p logs

deps: ## Install Go dependencies inside container
	@echo "Installing dependencies..."
	docker compose run --rm app go mod download
	docker compose run --rm app go mod tidy

db-up: ## Start database and redis only
	@echo "Starting database services..."
	docker compose up -d postgres redis
	@echo "Waiting for services to be healthy..."
	@docker compose exec postgres pg_isready -U billing_user -d billing_engine
	@docker compose exec redis redis-cli ping
	@echo "✅ Database ready at localhost:5432"
	@echo "✅ Redis ready at localhost:6379"

db-down: ## Stop all services
	docker compose down

server: deps ## Run server inside container with hot reload
	@echo "Starting server with hot reload in container..."
	docker compose run --rm -p 8080:8080 app sh -c "go install github.com/air-verse/air@latest && air -c .air.toml"

server-bg: deps ## Run server in background (daemon mode)
	@echo "Starting server in background..."
	docker compose up -d app

scheduler: deps ## Run scheduler inside container
	@echo "Starting scheduler in container..."
	docker compose --profile scheduler up scheduler

scheduler-bg: deps ## Run scheduler in background
	@echo "Starting scheduler in background..."
	docker compose --profile scheduler up -d scheduler

dev: db-up ## Start everything for development
	@echo "💡 Starting complete development environment..."
	@echo "💡 Run 'make server' to start the API server"
	@echo "💡 Run 'make scheduler-bg' if you need the scheduler"

dev-full: db-up deps ## Start everything including server and scheduler
	@echo "Starting complete development stack..."
	docker compose --profile scheduler up -d
    		
test: ## Run tests with coverage
	@echo "Running unit tests with coverage..."
	docker compose run --rm app sh -c "\
		go test -v ./... \
		-coverpkg=./... \
		-coverprofile=coverage.out && \
		go tool cover -func=coverage.out"

build: deps ## Build the application inside container
	@echo "Building application..."
	docker compose run --rm app go build -o bin/server cmd/server/main.go
	docker compose run --rm app go build -o bin/scheduler cmd/scheduler/main.go

logs: ## Show logs from running services
	docker compose logs -f

logs-app: ## Show only app logs
	docker compose logs -f app

shell: ## Get shell access to app container
	docker compose run --rm app sh

clean: ## Clean up everything
	@echo "Cleaning up..."
	docker compose --profile scheduler down -v
	docker compose down -v
	docker volume prune -f
	rm -rf logs/*.log

demo: ## Run demo (assumes server is running)
	@echo "Testing server health..."
	@curl -s http://localhost:8080/health || echo "❌ Server not responding. Start with: make server-bg"

status: ## Show status of all services
	docker compose ps

restart: ## Restart all services
	docker compose restart

# Development workflow shortcuts
quick-start: setup db-up server-bg ## Quick start: setup + db + server in background
	@echo "✅ Development environment ready!"
	@echo "🔗 Server: http://localhost:8080"
	@echo "🔍 Health: http://localhost:8080/health"

stop: ## Stop all running services
	docker compose --profile scheduler stop
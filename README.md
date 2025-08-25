# Billing Engine
Simple Loan Billing Engine in Go – calculates weekly installments, tracks outstanding balance, and detects delinquency.

## Quick Start

```bash
# 1. Setup project
make setup

# 2. Start complete development environment
make quick-start

# 3. Test the server
curl http://localhost:8080/health
```

## Alternative Start Methods

```bash
# Method 1: Step by step
make setup         # Setup project files
make db-up         # Start database services
make server        # Start server with hot reload (foreground)

# Method 2: Background services
make setup
make db-up
make server-bg     # Start server in background
make scheduler-bg  # Start scheduler in background (optional)

# Method 3: Everything at once
make dev-full      # Starts DB, server, and scheduler in background
```

## Problem Solved

**Billing Engine** with:
- ✅ Loan schedule generation (50 weeks × Rp 110,000)
- ✅ Outstanding amount calculation
- ✅ Delinquency detection (2+ missed payments)
- ✅ Payment processing

## API Endpoints

```bash
# Create loan
curl -X POST http://localhost:8080/api/v1/loans \
  -H "Content-Type: application/json" \
  -d '{"amount":1500,"duration_weeks":30,"interest_rate":0.12, "loan_id":"custom-loan-id"}'

# Get outstanding
curl http://localhost:8080/api/v1/loans/{id}/outstanding

# Check delinquency  
curl http://localhost:8080/api/v1/loans/{id}/delinquent

# Make payment
curl -X POST http://localhost:8080/api/v1/loans/{id}/payment \
  -H "Content-Type: application/json" \
  -d '{"amount": 110000, "loan_id":"custom-loan-id"}'
```

## Business Rules

- **Loan**: Rp 5,000,000 + 10% interest = Rp 5,500,000
- **Weekly Payment**: Rp 110,000 (exact amount only)
- **Duration**: 50 weeks
- **Delinquent**: 2+ consecutive missed payments

## Architecture

- **Language**: Go 1.24 (running in Docker)
- **Database**: PostgreSQL
- **Cache**: Redis
- **Router**: Gorilla Mux
- **Money**: Decimal precision (no floats!)
- **Testing**: Comprehensive test suite

## Development Commands

```bash
# 🚀 Development
make quick-start   # Setup + DB + Server (fastest way to start)
make dev           # Start DB services only
make dev-full      # Start everything in background

# 🔧 Services
make server        # Run server with hot reload (foreground)
make server-bg     # Run server in background
make scheduler     # Run scheduler (foreground)
make scheduler-bg  # Run scheduler in background

# 🗄️ Database
make db-up         # Start PostgreSQL and Redis
make db-down       # Stop all services

# 🧪 Testing
make test          # Run tests in container
make test-race     # Run tests with race detection

# 📦 Dependencies
make deps          # Install Go dependencies in container
make build         # Build binaries in container

# 🔍 Monitoring
make logs          # Show all service logs
make logs-app      # Show only app logs
make status        # Show service status
make demo          # Test server health

# 🛠️ Utilities
make shell         # Get shell access to app container
make clean         # Clean up everything
make restart       # Restart all services
```

## Container Architecture

All Go code runs inside Docker containers:

| Service | Container | Purpose |
|---------|-----------|---------|
| `app` | billing_app | Main API server |
| `scheduler` | billing_scheduler | Background jobs | *COMING SOON* |
| `postgres` | billing_db | Database |
| `redis` | billing_redis | Cache |

## Tech Stack

| Component | Library | Why |
|-----------|---------|-----|
| Runtime | Docker + Go 1.21 | Containerized development |
| HTTP | gorilla/mux | Simple, reliable |
| Database | sqlx + PostgreSQL | No ORM overhead |
| Cache | go-redis | Performance |
| Money | shopspring/decimal | Precision |
| Config | viper | Environment management |
| Validation | validator/v10 | Input safety |
| Cron | robfig/cron | Reliable scheduling |

## Development Workflow

1. **First time setup**:
   ```bash
   make quick-start
   ```

2. **Daily development**:
   ```bash
   make server        # Hot reload development
   # OR
   make server-bg     # Background server
   ```

3. **Testing**:
   ```bash
   make test
   ```

4. **Clean up**:
   ```bash
   make clean
   ```

## Environment Variables

All configuration is handled via environment variables in `.env` file:

- **DB_HOST**: `postgres` (Docker service name)
- **REDIS_HOST**: `redis` (Docker service name)  
- **SERVER_HOST**: `0.0.0.0` (bind to all interfaces in container)

## Implementation Highlights

1. **Containerized Development**: Everything runs in Docker
2. **Domain-Driven Design**: Clean separation of concerns
3. **Financial Precision**: Decimal arithmetic for money
4. **Service Health Checks**: Proper dependency waiting
5. **Hot Reload**: Air for development productivity
6. **Testing**: Comprehensive test coverage in containers
7. **Simple Architecture**: Easy to understand and extend

## Troubleshooting

**Server not starting?**
```bash
make logs-app      # Check app logs
make status        # Check service status
```

**Database connection issues?**
```bash
make db-down && make db-up    # Restart DB services
```

**Port conflicts?**
```bash
make clean         # Clean everything
lsof -i :8080      # Check what's using port 8080
```
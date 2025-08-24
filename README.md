# Billing Engine
Simple Loan Billing Engine in Go – calculates weekly installments, tracks outstanding balance, and detects delinquency.
## Quick Start

```bash
# 1. Setup
make setup

# 2. Start database
make db-up  

# 3. Start server (in another terminal)
make server

# 4. Test
curl http://localhost:8080/health
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
curl -X POST http://localhost:8080/loans \
  -H "Content-Type: application/json" \
  -d '{"amount": 5000000, "customer_id": "CUST001"}'

# Get outstanding
curl http://localhost:8080/loans/{id}/outstanding

# Check delinquency  
curl http://localhost:8080/loans/{id}/delinquent

# Make payment
curl -X POST http://localhost:8080/loans/{id}/payments \
  -H "Content-Type: application/json" \
  -d '{"amount": 110000}'
```

## Business Rules

- **Loan**: Rp 5,000,000 + 10% interest = Rp 5,500,000
- **Weekly Payment**: Rp 110,000 (exact amount only)
- **Duration**: 50 weeks
- **Delinquent**: 2+ consecutive missed payments

## Architecture

- **Language**: Go 1.21
- **Database**: PostgreSQL
- **Cache**: Redis
- **Router**: Gorilla Mux
- **Money**: Decimal precision (no floats!)
- **Testing**: Comprehensive test suite

## Development

```bash
# Hot reload server
make server

# Run tests
make test

# Run scheduler (if needed)
make scheduler

# Clean up
make clean
```

## Tech Stack

| Component | Library | Why |
|-----------|---------|-----|
| HTTP | gorilla/mux | Simple, reliable |
| Database | sqlx + PostgreSQL | No ORM overhead |
| Cache | go-redis | Performance |
| Money | shopspring/decimal | Precision |
| Config | viper | Environment management |
| Validation | validator/v10 | Input safety |
| Cron | robfig/cron | Reliable scheduling |

## Implementation Highlights

1. **Domain-Driven Design**: Clean separation of concerns
2. **Financial Precision**: Decimal arithmetic for money
3. **Validation**: Strict input validation
4. **Testing**: Comprehensive test coverage
5. **Simple Architecture**: Easy to understand and extend

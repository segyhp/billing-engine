package service

import (
	"context"

	"github.com/segyhp/billing-engine/internal/config"
	"github.com/segyhp/billing-engine/internal/domain"
	"github.com/segyhp/billing-engine/internal/repository"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

type BillingService struct {
	loanRepo    repository.LoanRepository
	paymentRepo repository.PaymentRepository
	redis       *redis.Client
	config      *config.Config
}

func NewBillingService(
	loanRepo repository.LoanRepository,
	paymentRepo repository.PaymentRepository,
	redis *redis.Client,
	config *config.Config,
) *BillingService {
	return &BillingService{
		loanRepo:    loanRepo,
		paymentRepo: paymentRepo,
		redis:       redis,
		config:      config,
	}
}

// CreateLoan creates a new loan with payment schedule
// TODO: Implement this method with business logic
func (s *BillingService) CreateLoan(ctx context.Context, loanID string) (*domain.Loan, []*domain.LoanSchedule, error) {
	// Business logic to implement:
	// 1. Create loan entity with the configured amount, interest rate, and duration
	// 2. Calculate weekly payment amount: (Principal + Interest) / Duration
	// 3. Generate payment schedule for 50 weeks
	// 4. Save loan and schedule to database
	// 5. Cache loan data in Redis for fast access

	panic("TODO: Implement CreateLoan business logic")
}

// GetOutstanding calculates and returns the outstanding balance for a loan
// TODO: Implement this method with business logic
func (s *BillingService) GetOutstanding(ctx context.Context, loanID string) (decimal.Decimal, error) {
	// Business logic to implement:
	// 1. Get loan details from database
	// 2. Get all payments made for this loan
	// 3. Calculate outstanding = Total Loan Amount - Sum of Payments
	// 4. Cache result in Redis with TTL
	// 5. Return outstanding amount

	panic("TODO: Implement GetOutstanding business logic")
}

// IsDelinquent checks if a borrower is delinquent (missed 2+ consecutive payments)
// TODO: Implement this method with business logic
func (s *BillingService) IsDelinquent(ctx context.Context, loanID string) (bool, int, error) {
	// Business logic to implement:
	// 1. Get loan schedule for the loan
	// 2. Get current week number based on loan start date
	// 3. Check which payments are overdue
	// 4. Count consecutive missed payments
	// 5. Return true if missed payments >= threshold (2 weeks)
	// 6. Cache delinquency status in Redis

	panic("TODO: Implement IsDelinquent business logic")
}

// MakePayment processes a payment for a loan
// TODO: Implement this method with business logic
func (s *BillingService) MakePayment(ctx context.Context, loanID string, amount decimal.Decimal) (*domain.Payment, error) {
	// Business logic to implement:
	// 1. Validate loan exists and is active
	// 2. Find the earliest unpaid week in the schedule
	// 3. Validate payment amount matches the weekly payment amount exactly
	// 4. Create payment record
	// 5. Update loan schedule status for that week
	// 6. Update cached outstanding balance
	// 7. Check if loan is fully paid and update status
	// 8. Return payment details

	panic("TODO: Implement MakePayment business logic")
}

// GetSchedule returns the payment schedule for a loan
// TODO: Implement this method with business logic
func (s *BillingService) GetSchedule(ctx context.Context, loanID string) ([]*domain.LoanSchedule, error) {
	// Business logic to implement:
	// 1. Get loan schedule from database
	// 2. Cache schedule in Redis for performance
	// 3. Return schedule with payment status for each week

	panic("TODO: Implement GetSchedule business logic")
}

// Helper method to calculate weekly payment amount
// TODO: Implement this helper method
func (s *BillingService) calculateWeeklyPayment() decimal.Decimal {
	// Business logic to implement:
	// Weekly Payment = (Principal + (Principal * Annual Interest Rate)) / Number of Weeks
	// Example: (5,000,000 + (5,000,000 * 0.10)) / 50 = 110,000

	panic("TODO: Implement calculateWeeklyPayment helper")
}

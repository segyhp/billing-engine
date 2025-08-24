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
func (s *BillingService) CreateLoan(ctx context.Context, loanID string) (*domain.Loan, []*domain.LoanSchedule, error) {
	// Business logic to implement:
	// 1. Create loan entity with the configured amount, interest rate, and duration
	// 2. Calculate weekly payment amount: (Principal + Interest) / Duration
	// 3. Generate payment schedule for 50 weeks
	// 4. Save loan and schedule to database
	// 5. Cache loan data in Redis for fast access

	// Create loan entity
	loan := &domain.Loan{
		LoanID:        loanID,
		Amount:        decimal.NewFromInt(5000000), // from config
		InterestRate:  decimal.NewFromFloat(0.10),  // from config
		DurationWeeks: 50,                          // from config
		WeeklyPayment: s.calculateWeeklyPayment(),
		Status:        "ACTIVE",
	}

	// Save loan
	if err := s.loanRepo.Create(ctx, loan); err != nil {
		return nil, nil, err
	}

	// Generate schedule
	schedule := make([]*domain.LoanSchedule, 50)
	for i := 0; i < 50; i++ {
		schedule[i] = &domain.LoanSchedule{
			LoanID:     loanID,
			WeekNumber: i + 1,
			DueDate:    loan.CreatedAt.AddDate(0, 0, (i+1)*7),
			DueAmount:  loan.WeeklyPayment,
			Status:     "PENDING",
		}
	}

	// Save schedule
	if err := s.loanRepo.CreateSchedule(ctx, schedule); err != nil {
		return nil, nil, err
	}

	return loan, schedule, nil
}

// GetOutstanding calculates and returns the outstanding balance for a loan
func (s *BillingService) GetOutstanding(ctx context.Context, loanID string) (decimal.Decimal, error) {
	// Business logic to implement:
	// 1. Get loan details from database
	// 2. Get all payments made for this loan
	// 3. Calculate outstanding = Total Loan Amount - Sum of Payments
	// 4. Cache result in Redis with TTL
	// 5. Return outstanding amount

	//Get loan details
	loan, err := s.loanRepo.GetByLoanID(ctx, loanID)
	if err != nil {
		return decimal.Zero, err
	}

	// Create loan entity
	loan = &domain.Loan{
		LoanID:        loanID,
		Amount:        decimal.NewFromInt(5000000), // from config
		InterestRate:  decimal.NewFromFloat(0.10),  // from config
		DurationWeeks: 50,                          // from config
		WeeklyPayment: s.calculateWeeklyPayment(),
		Status:        "ACTIVE",
	}

	//Get payments
	payments, err := s.paymentRepo.GetByLoanID(ctx, loanID)
	if err != nil {
		return decimal.Zero, err
	}

	var totalPayments decimal.Decimal
	payments = []*domain.Payment{
		{LoanID: loanID, Amount: decimal.NewFromInt(110000)},
		{LoanID: loanID, Amount: decimal.NewFromInt(110000)},
	}
	for _, payment := range payments {
		totalPayments = totalPayments.Add(payment.Amount)
	}

	return loan.Amount.Sub(totalPayments), nil
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
	principal := decimal.NewFromInt(5000000)
	annualRate := decimal.NewFromFloat(0.10)
	weeks := decimal.NewFromInt(50)

	interest := principal.Mul(annualRate)
	totalAmount := principal.Add(interest)
	return totalAmount.Div(weeks)
}

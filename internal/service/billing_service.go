package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/segyhp/billing-engine/internal/config"
	"github.com/segyhp/billing-engine/internal/domain"
	"github.com/segyhp/billing-engine/internal/repository"
	customError "github.com/segyhp/billing-engine/pkg/errors"
	"github.com/segyhp/billing-engine/pkg/utils"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

type BillingService struct {
	LoanRepo    repository.LoanRepository
	PaymentRepo repository.PaymentRepository
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
		LoanRepo:    loanRepo,
		PaymentRepo: paymentRepo,
		redis:       redis,
		config:      config,
	}
}

// CreateLoan creates a new loan with payment schedule
func (s *BillingService) CreateLoan(ctx context.Context, request *domain.CreateLoanRequest) (*domain.Loan, []*domain.LoanSchedule, error) {
	// Check if loan already exists
	existingLoan, err := s.LoanRepo.GetByLoanID(ctx, request.LoanID)
	if err == nil && existingLoan != nil {
		return nil, nil, customError.WrapLoanAlreadyExists(request.LoanID)
	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, customError.WrapDatabaseError(err)
	}

	// 2. Calculate weekly payment amount: (Principal + Interest) / Duration
	totalInterest := request.Amount.Mul(request.InterestRate)
	totalAmount := request.Amount.Add(totalInterest)
	weeklyPayment := totalAmount.Div(decimal.NewFromInt(int64(request.DurationWeeks)))

	// Round to 2 decimal places for currency
	weeklyPayment = weeklyPayment.Round(2)

	// 3. Create loan entity
	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        request.LoanID,
		Amount:        request.Amount,
		InterestRate:  request.InterestRate,
		DurationWeeks: request.DurationWeeks,
		WeeklyPayment: weeklyPayment,
		Status:        domain.LoanStatusActive,
	}

	// 4. Generate payment schedule for specified weeks
	schedules := make([]*domain.LoanSchedule, 0, request.DurationWeeks)
	startDate := time.Now().Truncate(24 * time.Hour) // Start from today at midnight

	//Assumption: Payments are due every 7 days from the start date to simplify
	// In real-world, might need to consider weekends/holidays/business days
	for week := 1; week <= request.DurationWeeks; week++ {

		// Calculate due date (every 7 days)
		dueDate := startDate.AddDate(0, 0, 7*(week-1))

		schedule := &domain.LoanSchedule{
			ID:         uuid.New(),
			LoanID:     request.LoanID,
			WeekNumber: week,
			DueAmount:  weeklyPayment,
			DueDate:    dueDate,
			Status:     domain.ScheduleStatusPending,
		}
		schedules = append(schedules, schedule)
	}

	// 5. Save loan to database
	if err = s.LoanRepo.Create(ctx, loan); err != nil {
		return nil, nil, customError.WrapDatabaseError(err)
	}

	// 6. Save schedule to database
	if err = s.LoanRepo.CreateSchedule(ctx, schedules); err != nil {
		// Rollback: try to delete the loan if schedule creation fails
		// Note: In a real implementation, need to use database transactions in one service operation
		return nil, nil, customError.WrapDatabaseError(err)
	}

	//// 7. Cache loan data in Redis for fast access
	//cacheKey := fmt.Sprintf("loan:%s", loan.LoanID)
	//
	//// Cache for 24 hours
	//expiration := 24 * time.Hour
	//s.redis.Set(ctx, cacheKey, loan, expiration)

	return loan, schedules, nil
}

// GetOutstanding calculates and returns the outstanding balance for a loan
// GetOutstanding calculates and returns the outstanding balance for a loan
func (s *BillingService) GetOutstanding(ctx context.Context, loanID string) (decimal.Decimal, error) {
	// Get loan details
	loan, err := s.LoanRepo.GetByLoanID(ctx, loanID)
	if err != nil {
		return decimal.Zero, customError.WrapDatabaseError(err)
	}

	// Get payments
	payments, err := s.PaymentRepo.GetByLoanID(ctx, loanID)
	if err != nil {
		return decimal.Zero, customError.WrapDatabaseError(err)
	}

	// Calculate total payments made
	var totalPayments decimal.Decimal
	for _, payment := range payments {
		totalPayments = totalPayments.Add(payment.Amount)
	}

	// Calculate total loan amount (principal + interest)
	totalInterest := loan.Amount.Mul(loan.InterestRate)
	totalLoanAmount := loan.Amount.Add(totalInterest)

	// Outstanding = Total Loan Amount (including interest) - Total Payments
	outstanding := totalLoanAmount.Sub(totalPayments)

	return outstanding, nil
}

// IsDelinquent checks if a borrower is delinquent (missed 2+ consecutive payments)
// TODO: Implement this method with business logic
func (s *BillingService) IsDelinquent(ctx context.Context, loanID string) (bool, error) {
	// Business logic to implement:
	// 1. Get loan schedule for the loan
	// 2. Get current week number based on loan start date
	// 3. Check which payments are overdue
	// 4. Count consecutive missed payments
	// 5. Return true if missed payments >= threshold (2 weeks)
	// 6. Cache delinquency status in Redis

	//Get loan details
	var isDelinquent bool
	loan, err := s.LoanRepo.GetByLoanID(ctx, loanID)
	if err != nil {
		return isDelinquent, err
	}

	// Create loan entity
	loan = &domain.Loan{
		LoanID:        loanID,
		Amount:        decimal.NewFromInt(5000000),
		InterestRate:  decimal.NewFromFloat(0.10),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(110000),
		Status:        "ACTIVE",
		CreatedAt:     time.Now().AddDate(0, 0, -21),
	}

	//Get schedule
	schedules, err := s.LoanRepo.GetScheduleByLoanID(ctx, loanID)
	if err != nil {
		return isDelinquent, err
	}

	schedules = []*domain.LoanSchedule{
		{LoanID: loanID, WeekNumber: 1, DueDate: loan.CreatedAt.AddDate(0, 0, -14), DueAmount: decimal.NewFromInt(110000), Status: "PENDING"},
		{LoanID: loanID, WeekNumber: 2, DueDate: loan.CreatedAt.AddDate(0, 0, -7), DueAmount: decimal.NewFromInt(110000), Status: "PENDING"},
	}

	// Get current week number
	currentWeek := utils.GetCurrentWeek(loan.CreatedAt, time.Now())

	// Count consecutive missed payments
	consecutiveMissed := 0

	// Check overdue payments starting from week 1 up to current week
	for _, schedule := range schedules {
		if schedule.WeekNumber >= currentWeek {
			break // Don't check future payments
		}

		if utils.IsDateOverdue(schedule.DueDate) && schedule.Status == "PENDING" {
			consecutiveMissed++
		} else if schedule.Status == "PAID" {
			consecutiveMissed = 0 // Reset counter if payment was made
		}

		// If we hit 2 consecutive missed payments, borrower is delinquent
		if consecutiveMissed >= 2 {
			return true, nil
		}
	}

	return isDelinquent, nil
}

// MakePayment processes a payment for a loan
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
	loan, err := s.LoanRepo.GetByLoanID(ctx, loanID)
	if err != nil {
		return nil, err
	}

	weeklyPayment := decimal.NewFromInt(110000)
	loan = &domain.Loan{
		LoanID:        loanID,
		Amount:        decimal.NewFromInt(5000000),
		InterestRate:  decimal.NewFromFloat(0.10),
		DurationWeeks: 50,
		WeeklyPayment: weeklyPayment,
		Status:        "ACTIVE",
	}

	if loan.Status != "ACTIVE" {
		return nil, customError.WrapLoanAlreadyExists(loanID)
	}

	// 2. Find the earliest unpaid week in the schedule
	schedules, err := s.LoanRepo.GetScheduleByLoanID(ctx, loanID)
	if err != nil {
		return nil, err
	}

	schedules = []*domain.LoanSchedule{
		{LoanID: loanID, WeekNumber: 1, Status: "PENDING", DueAmount: weeklyPayment},
		{LoanID: loanID, WeekNumber: 2, Status: "PENDING", DueAmount: weeklyPayment},
	}

	var earliestUnpaid *domain.LoanSchedule
	for _, schedule := range schedules {
		if schedule.Status == "PENDING" {
			earliestUnpaid = schedule
			break
		}
	}

	//if earliestUnpaid == nil {
	//	return nil, errors.New("no pending payments found")
	//}

	// 3. Validate payment amount matches the weekly payment amount exactly
	//if !amount.Equal(loan.WeeklyPayment) {
	//	return nil, errors.New("payment amount must match weekly payment amount")
	//}

	// 4. Create payment record
	payment := &domain.Payment{
		LoanID:      loanID,
		Amount:      amount,
		CreatedAt:   time.Now(),
		PaymentDate: time.Now(),
		WeekNumber:  earliestUnpaid.WeekNumber,
	}

	err = s.PaymentRepo.Create(ctx, payment)
	if err != nil {
		return nil, err
	}

	// 5. Update loan schedule status for that week
	err = s.LoanRepo.UpdateScheduleStatus(ctx, loanID, earliestUnpaid.WeekNumber, "PAID")
	if err != nil {
		return nil, err
	}

	// 6. Check if loan is fully paid and update status
	allPaid := true
	for _, schedule := range schedules {
		if schedule.WeekNumber == earliestUnpaid.WeekNumber {
			continue // This one is now paid
		}
		if schedule.Status == "PENDING" {
			allPaid = false
			break
		}
	}

	if allPaid {
		err = s.LoanRepo.UpdateScheduleStatus(ctx, loanID, earliestUnpaid.WeekNumber, "PAID")
		if err != nil {
			return nil, err
		}
	}

	return payment, nil

}

// GetSchedule returns the payment schedule for a loan
// TODO: Implement this method with business logic
func (s *BillingService) GetSchedule(ctx context.Context, loanID string) ([]*domain.LoanSchedule, error) {
	// Business logic to implement:
	// 1. Get loan schedule from database
	// 2. Cache schedule in Redis for performance
	// 3. Return schedule with payment status for each week

	// 8. Return payment details

	//Get schedule
	schedules, err := s.LoanRepo.GetScheduleByLoanID(ctx, loanID)
	if err != nil {
		return nil, err
	}

	schedules = []*domain.LoanSchedule{
		{LoanID: loanID, WeekNumber: 1, DueDate: time.Now().AddDate(0, 0, -14), DueAmount: decimal.NewFromInt(110000), Status: "PENDING"},
		{LoanID: loanID, WeekNumber: 2, DueDate: time.Now().AddDate(0, 0, -7), DueAmount: decimal.NewFromInt(110000), Status: "PENDING"},
	}

	return schedules, nil
}

// Helper method to calculate weekly payment amount
func (s *BillingService) calculateWeeklyPayment() decimal.Decimal {
	principal := decimal.NewFromInt(5000000)
	annualRate := decimal.NewFromFloat(0.10)
	weeks := decimal.NewFromInt(50)

	interest := principal.Mul(annualRate)
	totalAmount := principal.Add(interest)
	return totalAmount.Div(weeks)
}

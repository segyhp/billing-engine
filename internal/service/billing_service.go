package service

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/segyhp/billing-engine/internal/config"
	"github.com/segyhp/billing-engine/internal/domain"
	"github.com/segyhp/billing-engine/internal/repository"
	customError "github.com/segyhp/billing-engine/pkg/errors"
	//"github.com/segyhp/billing-engine/pkg/utils"

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
func (s *BillingService) IsDelinquent(ctx context.Context, loanID string) (bool, error) {
	// Get loan details
	loan, err := s.LoanRepo.GetByLoanID(ctx, loanID)
	if err != nil {
		return false, customError.WrapDatabaseError(err)
	}

	if loan.Status != domain.LoanStatusActive {
		// Only active loans can be delinquent
		return false, customError.WrapLoanAlreadyClosed(loanID)
	}

	// Get loan schedule for the loan
	schedules, err := s.LoanRepo.GetScheduleByLoanID(ctx, loanID)
	if err != nil {
		return false, customError.WrapDatabaseError(err)
	}

	// Sort schedules by due date to ensure proper order
	sort.Slice(schedules, func(i, j int) bool {
		return schedules[i].DueDate.Before(schedules[j].DueDate)
	})

	// Count consecutive missed payments
	consecutiveMissed := 0
	now := time.Now()
	const threshold = 2 // 2 weeks threshold

	// Check which payments are overdue
	for _, schedule := range schedules {
		// Only check past due dates (not including today)
		// for now we assume that timezone is not an issue
		// In real-world, need to consider timezone differences between server and client (vary in timezone)
		// e.g., if due date is today but time has not yet reached due time
		if schedule.DueDate.After(now.Truncate(24 * time.Hour)) {
			break // Don't check future payments or today's payment
		}

		// Check if this payment is overdue (past due date and still pending)
		if schedule.Status == domain.ScheduleStatusPending {
			consecutiveMissed++

			// Return true if missed payments >= threshold (2 weeks)
			if consecutiveMissed >= threshold {
				return true, nil
			}
		} else if schedule.Status == domain.ScheduleStatusPaid {
			// Reset counter when payment is made
			consecutiveMissed = 0
		}
		// Note: We don't reset for future payments - only process overdue ones
	}

	return false, nil
}

// MakePayment processes a payment for a loan
// MakePayment processes a payment for a loan
func (s *BillingService) MakePayment(ctx context.Context, request domain.MakePaymentRequest) (*domain.Payment, error) {
	// 1. Validate payment amount
	if request.Amount.LessThanOrEqual(decimal.Zero) {
		invalidAmount, _ := request.Amount.Float64()
		return nil, customError.WrapInvalidPaymentAmount(invalidAmount)
	}

	// 2. Validate loan exists and is active
	loan, err := s.LoanRepo.GetByLoanID(ctx, request.LoanID)
	if err != nil {
		return nil, customError.WrapDatabaseError(err)
	}

	if loan.Status != domain.LoanStatusActive {
		return nil, customError.WrapLoanAlreadyClosed(request.LoanID)
	}

	// 3. Find the earliest unpaid week in the schedule
	schedules, err := s.LoanRepo.GetScheduleByLoanID(ctx, request.LoanID)
	if err != nil {
		return nil, customError.WrapDatabaseError(err)
	}

	// Find the earliest unpaid week
	var earliestUnpaid *domain.LoanSchedule
	for _, schedule := range schedules {
		if schedule.Status == domain.ScheduleStatusPending {
			earliestUnpaid = schedule
			break
		}
	}

	if earliestUnpaid == nil {
		return nil, customError.WrapNoOutstandingBalance(request.LoanID)
	}

	// 4. Validate payment amount matches exactly
	if !request.Amount.Equal(loan.WeeklyPayment) {
		invalidAmount, _ := request.Amount.Float64()
		return nil, customError.WrapInvalidPaymentAmount(invalidAmount)
	}

	// 5. Create payment record
	payment := &domain.Payment{
		ID:          uuid.New(),
		LoanID:      request.LoanID,
		Amount:      request.Amount,
		PaymentDate: time.Now(),
		WeekNumber:  earliestUnpaid.WeekNumber,
	}

	err = s.PaymentRepo.Create(ctx, payment)
	if err != nil {
		return nil, customError.WrapDatabaseError(err)
	}

	// 6. Update loan schedule status for that week
	err = s.LoanRepo.UpdateScheduleStatus(ctx, request.LoanID, earliestUnpaid.WeekNumber, "PAID")
	if err != nil {
		return nil, customError.WrapDatabaseError(err)
	}

	// 7. Check if loan is fully paid and update status
	allPaid := true
	for _, schedule := range schedules {
		// Skip the schedule we just paid
		if schedule.WeekNumber == earliestUnpaid.WeekNumber {
			continue
		}
		// Check if any other schedule is still pending
		if schedule.Status == domain.ScheduleStatusPending {
			allPaid = false
			break
		}
	}

	if allPaid {
		loan.Status = domain.LoanStatusClosed
		err = s.LoanRepo.Update(ctx, loan)
		if err != nil {
			return nil, customError.WrapDatabaseError(err)
		}
	}

	return payment, nil
}

// Helper function to calculate current week number from loan start date
func (s *BillingService) getCurrentWeekFromLoanStart(startDate, currentDate time.Time) int {
	if currentDate.Before(startDate) {
		return 0
	}

	duration := currentDate.Sub(startDate)
	weeks := int(duration.Hours() / (24 * 7)) // Convert to weeks
	return weeks + 1                          // Week 1 starts from loan start date
}

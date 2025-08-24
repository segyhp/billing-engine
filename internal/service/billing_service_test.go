package service

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/segyhp/billing-engine/internal/domain"
	"github.com/segyhp/billing-engine/tests/mocks"
)

func TestCreateLoan_Success(t *testing.T) {
	// Arrange
	mockLoanRepo := &mocks.MockLoanRepository{}
	mockPaymentRepo := &mocks.MockPaymentRepository{}

	service := &BillingService{
		loanRepo:    mockLoanRepo,
		paymentRepo: mockPaymentRepo,
		// config with: Amount: 5000000, Rate: 0.10, Duration: 50
	}

	loanID := "LOAN123"

	// Set expectations
	mockLoanRepo.On("Create", mock.Anything, mock.MatchedBy(func(loan *domain.Loan) bool {
		return loan.LoanID == loanID
	})).Return(nil)

	mockLoanRepo.On("CreateSchedule", mock.Anything, mock.MatchedBy(func(schedules []*domain.LoanSchedule) bool {
		return len(schedules) == 50
	})).Return(nil)

	// Act
	loan, schedule, err := service.CreateLoan(context.Background(), loanID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, loanID, loan.LoanID)
	assert.Equal(t, 50, len(schedule))
	assert.True(t, loan.WeeklyPayment.Equal(decimal.NewFromInt(110000)))

	mockLoanRepo.AssertExpectations(t)
}

func TestGetOutstanding_Success(t *testing.T) {
	mockLoanRepo := &mocks.MockLoanRepository{}
	mockPaymentRepo := &mocks.MockPaymentRepository{}

	service := &BillingService{
		loanRepo:    mockLoanRepo,
		paymentRepo: mockPaymentRepo,
	}

	loanID := "LOAN123"
	loan := &domain.Loan{
		LoanID:        loanID,
		Amount:        decimal.NewFromInt(5000000),
		InterestRate:  decimal.NewFromFloat(0.10),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(110000),
		Status:        "ACTIVE",
	}

	payments := []*domain.Payment{
		{LoanID: loanID, Amount: decimal.NewFromInt(110000)},
		{LoanID: loanID, Amount: decimal.NewFromInt(110000)},
	}

	mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
	mockPaymentRepo.On("GetByLoanID", mock.Anything, loanID).Return(payments, nil)

	// Act
	outstanding, err := service.GetOutstanding(context.Background(), loanID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, outstanding.Equal(decimal.NewFromInt(4780000))) // 5000000 + 500000 - 220000

	mockLoanRepo.AssertExpectations(t)
	mockPaymentRepo.AssertExpectations(t)
}

func TestIsDelinquent_Success(t *testing.T) {
	mockLoanRepo := &mocks.MockLoanRepository{}
	mockPaymentRepo := &mocks.MockPaymentRepository{}

	service := &BillingService{
		loanRepo:    mockLoanRepo,
		paymentRepo: mockPaymentRepo,
	}

	loanID := "LOAN123"
	loan := &domain.Loan{
		LoanID:        loanID,
		Amount:        decimal.NewFromInt(5000000),
		InterestRate:  decimal.NewFromFloat(0.10),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(110000),
		Status:        "ACTIVE",
		CreatedAt:     time.Now(),
	}

	schedules := []*domain.LoanSchedule{
		{LoanID: loanID, WeekNumber: 1, DueDate: loan.CreatedAt.AddDate(0, 0, -14), DueAmount: decimal.NewFromInt(110000), Status: "PENDING"},
		{LoanID: loanID, WeekNumber: 2, DueDate: loan.CreatedAt.AddDate(0, 0, -7), DueAmount: decimal.NewFromInt(110000), Status: "PENDING"},
	}

	mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
	mockLoanRepo.On("GetScheduleByLoanID", mock.Anything, loanID).Return(schedules, nil)

	// Act
	delinquent, err := service.IsDelinquent(context.Background(), loanID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, delinquent)

	mockLoanRepo.AssertExpectations(t)
	mockPaymentRepo.AssertExpectations(t)
}

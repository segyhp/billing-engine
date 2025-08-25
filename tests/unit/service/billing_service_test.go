package service

import (
	"context"
	"testing"
	"time"

	billingService "github.com/segyhp/billing-engine/internal/service"
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

	service := &billingService.BillingService{
		LoanRepo:    mockLoanRepo,
		PaymentRepo: mockPaymentRepo,
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

	service := &billingService.BillingService{
		LoanRepo:    mockLoanRepo,
		PaymentRepo: mockPaymentRepo,
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

	service := &billingService.BillingService{
		LoanRepo:    mockLoanRepo,
		PaymentRepo: mockPaymentRepo,
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

func TestMakePayment_Success(t *testing.T) {
	mockLoanRepo := &mocks.MockLoanRepository{}
	mockPaymentRepo := &mocks.MockPaymentRepository{}

	service := &billingService.BillingService{
		LoanRepo:    mockLoanRepo,
		PaymentRepo: mockPaymentRepo,
	}

	loanID := "LOAN123"
	weeklyPayment := decimal.NewFromInt(110000)

	loan := &domain.Loan{
		LoanID:        loanID,
		Amount:        decimal.NewFromInt(5000000),
		InterestRate:  decimal.NewFromFloat(0.10),
		DurationWeeks: 50,
		WeeklyPayment: weeklyPayment,
		Status:        "ACTIVE",
	}

	schedules := []*domain.LoanSchedule{
		{LoanID: loanID, WeekNumber: 1, Status: "PENDING", DueAmount: weeklyPayment},
		{LoanID: loanID, WeekNumber: 2, Status: "PENDING", DueAmount: weeklyPayment},
	}

	mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
	mockLoanRepo.On("GetScheduleByLoanID", mock.Anything, loanID).Return(schedules, nil)
	mockPaymentRepo.On("Create", mock.Anything, mock.MatchedBy(func(payment *domain.Payment) bool {
		return payment.LoanID == loanID && payment.Amount.Equal(weeklyPayment)
	})).Return(nil)
	mockLoanRepo.On("UpdateScheduleStatus", mock.Anything, loanID, 1, "PAID").Return(nil)

	// Act
	payment, err := service.MakePayment(context.Background(), loanID, weeklyPayment)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, loanID, payment.LoanID)
	assert.True(t, payment.Amount.Equal(weeklyPayment))

	mockLoanRepo.AssertExpectations(t)
	mockPaymentRepo.AssertExpectations(t)
}

//func TestMakePayment_InvalidAmount(t *testing.T) {
//	mockLoanRepo := &mocks.MockLoanRepository{}
//	mockPaymentRepo := &mocks.MockPaymentRepository{}
//
//	service := &BillingService{
//		loanRepo:    mockLoanRepo,
//		paymentRepo: mockPaymentRepo,
//	}
//
//	loanID := "LOAN123"
//	loan := &domain.Loan{
//		LoanID:        loanID,
//		WeeklyPayment: decimal.NewFromInt(110000),
//		Status:        "ACTIVE",
//	}
//
//	schedules := []*domain.LoanSchedule{
//		{LoanID: loanID, WeekNumber: 1, Status: "PENDING"},
//	}
//
//	mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
//	mockLoanRepo.On("GetScheduleByLoanID", mock.Anything, loanID).Return(schedules, nil)
//
//	// Act - try to pay wrong amount
//	wrongAmount := decimal.NewFromInt(50000)
//	payment, err := service.MakePayment(context.Background(), loanID, wrongAmount)
//
//	// Assert
//	assert.Error(t, err)
//	assert.Nil(t, payment)
//	assert.Contains(t, err.Error(), "payment amount must match weekly payment amount")
//
//	mockLoanRepo.AssertExpectations(t)
//}
//
//func TestMakePayment_LoanNotActive(t *testing.T) {
//	mockLoanRepo := &mocks.MockLoanRepository{}
//	mockPaymentRepo := &mocks.MockPaymentRepository{}
//
//	service := &BillingService{
//		loanRepo:    mockLoanRepo,
//		paymentRepo: mockPaymentRepo,
//	}
//
//	loanID := "LOAN123"
//	loan := &domain.Loan{
//		LoanID: loanID,
//		Status: "PAID", // Not active
//	}
//
//	mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
//
//	// Act
//	payment, err := service.MakePayment(context.Background(), loanID, decimal.NewFromInt(110000))
//
//	// Assert
//	assert.Error(t, err)
//	assert.Nil(t, payment)
//	assert.Contains(t, err.Error(), "loan is not active")
//
//	mockLoanRepo.AssertExpectations(t)
//}

func TestBillingService_GetSchedule(t *testing.T) {
	mockLoanRepo := &mocks.MockLoanRepository{}
	mockPaymentRepo := &mocks.MockPaymentRepository{}

	service := &billingService.BillingService{
		LoanRepo:    mockLoanRepo,
		PaymentRepo: mockPaymentRepo,
	}
	loanID := "LOAN123"
	weeklyPayment := decimal.NewFromInt(110000)

	schedules := []*domain.LoanSchedule{
		{LoanID: loanID, WeekNumber: 1, Status: "PENDING", DueAmount: weeklyPayment},
		{LoanID: loanID, WeekNumber: 2, Status: "PENDING", DueAmount: weeklyPayment},
	}

	mockLoanRepo.On("GetScheduleByLoanID", mock.Anything, loanID).Return(schedules, nil)

	// Act
	payment, err := service.GetSchedule(context.Background(), loanID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, schedules[0].LoanID, loanID)

	mockLoanRepo.AssertExpectations(t)
	mockPaymentRepo.AssertExpectations(t)
}

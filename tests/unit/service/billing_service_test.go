package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	billingService "github.com/segyhp/billing-engine/internal/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/segyhp/billing-engine/internal/domain"
	"github.com/segyhp/billing-engine/tests/mocks"
)

func TestCreateLoan(t *testing.T) {
	tests := []struct {
		name           string
		loanID         string
		amount         decimal.Decimal
		interestRate   decimal.Decimal
		durationWeeks  int
		setupMocks     func(*mocks.MockLoanRepository, *mocks.MockPaymentRepository, string)
		expectedError  bool
		errorContains  string
		validateResult func(*testing.T, *domain.Loan, []*domain.LoanSchedule)
	}{
		{
			name:          "Success - Create new loan",
			loanID:        "LOAN123",
			amount:        decimal.NewFromFloat(5000000),
			interestRate:  decimal.NewFromFloat(0.10),
			durationWeeks: 50,
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(nil, sql.ErrNoRows)
				mockLoanRepo.On("Create", mock.Anything, mock.MatchedBy(func(loan *domain.Loan) bool {
					return loan.LoanID == loanID
				})).Return(nil)
				mockLoanRepo.On("CreateSchedule", mock.Anything, mock.MatchedBy(func(schedules []*domain.LoanSchedule) bool {
					return len(schedules) == 50
				})).Return(nil)
			},
			expectedError: false,
			validateResult: func(t *testing.T, loan *domain.Loan, schedule []*domain.LoanSchedule) {
				assert.Equal(t, "LOAN123", loan.LoanID)
				assert.Equal(t, 50, len(schedule))
				assert.True(t, loan.WeeklyPayment.Equal(decimal.NewFromInt(110000)))
			},
		},
		{
			name:          "Failure - Loan already exists",
			loanID:        "LOAN456",
			amount:        decimal.NewFromFloat(5000000),
			interestRate:  decimal.NewFromFloat(0.10),
			durationWeeks: 50,
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				existingLoan := &domain.Loan{LoanID: loanID}
				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(existingLoan, nil)
			},
			expectedError: true,
			errorContains: "already exists",
			validateResult: func(t *testing.T, loan *domain.Loan, schedule []*domain.LoanSchedule) {
				assert.Nil(t, loan)
				assert.Nil(t, schedule)
			},
		},
		{
			name:          "Failure - Database error on GetByLoanID",
			loanID:        "LOAN789",
			amount:        decimal.NewFromFloat(5000000),
			interestRate:  decimal.NewFromFloat(0.10),
			durationWeeks: 50,
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(nil, errors.New("database connection error"))
			},
			expectedError: true,
			errorContains: "database",
			validateResult: func(t *testing.T, loan *domain.Loan, schedule []*domain.LoanSchedule) {
				assert.Nil(t, loan)
				assert.Nil(t, schedule)
			},
		},
		{
			name:          "Failure - Database error on Create loan",
			loanID:        "LOAN101",
			amount:        decimal.NewFromFloat(5000000),
			interestRate:  decimal.NewFromFloat(0.10),
			durationWeeks: 50,
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(nil, sql.ErrNoRows)
				mockLoanRepo.On("Create", mock.Anything, mock.MatchedBy(func(loan *domain.Loan) bool {
					return loan.LoanID == loanID
				})).Return(errors.New("failed to create loan"))
			},
			expectedError: true,
			errorContains: "database",
			validateResult: func(t *testing.T, loan *domain.Loan, schedule []*domain.LoanSchedule) {
				assert.Nil(t, loan)
				assert.Nil(t, schedule)
			},
		},
		{
			name:          "Failure - Database error on CreateSchedule",
			loanID:        "LOAN202",
			amount:        decimal.NewFromFloat(5000000),
			interestRate:  decimal.NewFromFloat(0.10),
			durationWeeks: 50,
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(nil, sql.ErrNoRows)
				mockLoanRepo.On("Create", mock.Anything, mock.MatchedBy(func(loan *domain.Loan) bool {
					return loan.LoanID == loanID
				})).Return(nil)
				mockLoanRepo.On("CreateSchedule", mock.Anything, mock.MatchedBy(func(schedules []*domain.LoanSchedule) bool {
					return len(schedules) == 50
				})).Return(errors.New("failed to create schedule"))
			},
			expectedError: true,
			errorContains: "database",
			validateResult: func(t *testing.T, loan *domain.Loan, schedule []*domain.LoanSchedule) {
				assert.Nil(t, loan)
				assert.Nil(t, schedule)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockLoanRepo := &mocks.MockLoanRepository{}
			mockPaymentRepo := &mocks.MockPaymentRepository{}

			service := &billingService.BillingService{
				LoanRepo:    mockLoanRepo,
				PaymentRepo: mockPaymentRepo,
			}

			tt.setupMocks(mockLoanRepo, mockPaymentRepo, tt.loanID)

			request := &domain.CreateLoanRequest{
				LoanID:        tt.loanID,
				Amount:        tt.amount,
				InterestRate:  tt.interestRate,
				DurationWeeks: tt.durationWeeks,
			}

			// Act
			loan, schedule, err := service.CreateLoan(context.Background(), request)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			tt.validateResult(t, loan, schedule)
			mockLoanRepo.AssertExpectations(t)
		})
	}
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

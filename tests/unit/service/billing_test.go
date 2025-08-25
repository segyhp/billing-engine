package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
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

			service := billingService.NewBillingService(mockLoanRepo, mockPaymentRepo, nil, nil)

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

func TestGetOutstanding(t *testing.T) {
	tests := []struct {
		name                string
		loanID              string
		setupMocks          func(*mocks.MockLoanRepository, *mocks.MockPaymentRepository, string)
		expectedError       bool
		errorContains       string
		expectedOutstanding decimal.Decimal
		expectedLoanStatus  string
	}{
		{
			name:   "Success - No payments made",
			loanID: "LOAN123",
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:        loanID,
					Amount:        decimal.NewFromInt(5000000), // Principal: 5,000,000
					InterestRate:  decimal.NewFromFloat(0.10),  // Interest rate: 10%
					DurationWeeks: 50,
					WeeklyPayment: decimal.NewFromInt(110000),
					Status:        domain.LoanStatusActive,
				}
				var payments []*domain.Payment // No payments

				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
				mockPaymentRepo.On("GetByLoanID", mock.Anything, loanID).Return(payments, nil)
			},
			expectedError:       false,
			expectedOutstanding: decimal.NewFromInt(5500000), // 5,000,000 + 500,000 (10% interest) - 0
		},
		{
			name:   "Success - Partial payments made",
			loanID: "LOAN123",
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:        loanID,
					Amount:        decimal.NewFromInt(5000000),
					InterestRate:  decimal.NewFromFloat(0.10),
					DurationWeeks: 50,
					WeeklyPayment: decimal.NewFromInt(110000),
					Status:        domain.LoanStatusActive,
				}
				payments := []*domain.Payment{
					{LoanID: loanID, Amount: decimal.NewFromInt(110000)},
					{LoanID: loanID, Amount: decimal.NewFromInt(110000)},
				}

				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
				mockPaymentRepo.On("GetByLoanID", mock.Anything, loanID).Return(payments, nil)
			},
			expectedError:       false,
			expectedOutstanding: decimal.NewFromInt(5280000), // 5,500,000 - 220,000
		},
		{
			name:   "Success - Loan fully paid",
			loanID: "LOAN123",
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:        loanID,
					Amount:        decimal.NewFromInt(5000000),
					InterestRate:  decimal.NewFromFloat(0.10),
					DurationWeeks: 50,
					WeeklyPayment: decimal.NewFromInt(110000),
					Status:        domain.LoanStatusClosed,
				}

				// Generate 50 payments of 110,000 each = 5,500,000 total
				payments := make([]*domain.Payment, 50)
				for i := 0; i < 50; i++ {
					payments[i] = &domain.Payment{
						LoanID: loanID,
						Amount: decimal.NewFromInt(110000),
					}
				}

				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
				mockPaymentRepo.On("GetByLoanID", mock.Anything, loanID).Return(payments, nil)
			},
			expectedError:       false,
			expectedOutstanding: decimal.Zero, // Fully paid
			expectedLoanStatus:  domain.LoanStatusClosed,
		},
		{
			name:   "Failure - Loan not found",
			loanID: "NONEXISTENT",
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(nil, sql.ErrNoRows)
			},
			expectedError:       true,
			errorContains:       "database",
			expectedOutstanding: decimal.Zero,
		},
		{
			name:   "Failure - Database error getting loan",
			loanID: "LOAN123",
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(nil, errors.New("database connection error"))
			},
			expectedError:       true,
			errorContains:       "database",
			expectedOutstanding: decimal.Zero,
		},
		{
			name:   "Failure - Database error getting payments",
			loanID: "LOAN123",
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:        loanID,
					Amount:        decimal.NewFromInt(5000000),
					InterestRate:  decimal.NewFromFloat(0.10),
					DurationWeeks: 50,
					WeeklyPayment: decimal.NewFromInt(110000),
					Status:        domain.LoanStatusActive,
				}

				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
				mockPaymentRepo.On("GetByLoanID", mock.Anything, loanID).Return(nil, errors.New("payment query failed"))
			},
			expectedError:       true,
			errorContains:       "database",
			expectedOutstanding: decimal.Zero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockLoanRepo := &mocks.MockLoanRepository{}
			mockPaymentRepo := &mocks.MockPaymentRepository{}

			service := billingService.NewBillingService(mockLoanRepo, mockPaymentRepo, nil, nil)

			tt.setupMocks(mockLoanRepo, mockPaymentRepo, tt.loanID)

			// Act
			outstanding, err := service.GetOutstanding(context.Background(), tt.loanID)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, outstanding.Equal(tt.expectedOutstanding),
					"Expected %s, got %s", tt.expectedOutstanding.String(), outstanding.String())
			}

			mockLoanRepo.AssertExpectations(t)
			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestIsDelinquent(t *testing.T) {
	tests := []struct {
		name               string
		loanID             string
		setupMocks         func(*mocks.MockLoanRepository, *mocks.MockPaymentRepository, string)
		expectedError      bool
		errorContains      string
		expectedDelinquent bool
	}{
		{
			name:   "Success - Loan is delinquent (2 consecutive missed payments)",
			loanID: "LOAN123",
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:        loanID,
					Amount:        decimal.NewFromInt(5000000),
					InterestRate:  decimal.NewFromFloat(0.10),
					DurationWeeks: 50,
					WeeklyPayment: decimal.NewFromInt(110000),
					Status:        domain.LoanStatusActive,
					CreatedAt:     time.Now().AddDate(0, 0, -21), // 3 weeks ago
				}

				schedules := []*domain.LoanSchedule{
					{LoanID: loanID, WeekNumber: 1, DueDate: time.Now().AddDate(0, 0, -14), DueAmount: decimal.NewFromInt(110000), Status: domain.ScheduleStatusPending},
					{LoanID: loanID, WeekNumber: 2, DueDate: time.Now().AddDate(0, 0, -7), DueAmount: decimal.NewFromInt(110000), Status: domain.ScheduleStatusPending},
					{LoanID: loanID, WeekNumber: 3, DueDate: time.Now(), DueAmount: decimal.NewFromInt(110000), Status: domain.ScheduleStatusPending},
				}

				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
				mockLoanRepo.On("GetScheduleByLoanID", mock.Anything, loanID).Return(schedules, nil)
			},
			expectedError:      false,
			expectedDelinquent: true,
		},
		{
			name:   "Success - Loan is not delinquent (only 1 missed payment)",
			loanID: "LOAN124",
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:        loanID,
					Amount:        decimal.NewFromInt(5000000),
					InterestRate:  decimal.NewFromFloat(0.10),
					DurationWeeks: 50,
					WeeklyPayment: decimal.NewFromInt(110000),
					Status:        domain.LoanStatusActive,
					CreatedAt:     time.Now().AddDate(0, 0, -14), // 2 weeks ago
				}

				schedules := []*domain.LoanSchedule{
					{LoanID: loanID, WeekNumber: 1, DueDate: time.Now().AddDate(0, 0, -7), DueAmount: decimal.NewFromInt(110000), Status: domain.ScheduleStatusPending},
					{LoanID: loanID, WeekNumber: 2, DueDate: time.Now().AddDate(0, 0, 7), DueAmount: decimal.NewFromInt(110000), Status: domain.ScheduleStatusPending}, // Future payment
				}

				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
				mockLoanRepo.On("GetScheduleByLoanID", mock.Anything, loanID).Return(schedules, nil)
			},
			expectedError:      false,
			expectedDelinquent: false,
		},
		{
			name:   "Success - Not delinquent (payments made on time)",
			loanID: "LOAN125",
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:        loanID,
					Amount:        decimal.NewFromInt(5000000),
					InterestRate:  decimal.NewFromFloat(0.10),
					DurationWeeks: 50,
					WeeklyPayment: decimal.NewFromInt(110000),
					Status:        domain.LoanStatusActive,
					CreatedAt:     time.Now().AddDate(0, 0, -21), // 3 weeks ago
				}

				schedules := []*domain.LoanSchedule{
					{LoanID: loanID, WeekNumber: 1, DueDate: time.Now().AddDate(0, 0, -14), DueAmount: decimal.NewFromInt(110000), Status: domain.ScheduleStatusPaid},
					{LoanID: loanID, WeekNumber: 2, DueDate: time.Now().AddDate(0, 0, -7), DueAmount: decimal.NewFromInt(110000), Status: domain.ScheduleStatusPaid},
					{LoanID: loanID, WeekNumber: 3, DueDate: time.Now(), DueAmount: decimal.NewFromInt(110000), Status: domain.ScheduleStatusPending},
				}

				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
				mockLoanRepo.On("GetScheduleByLoanID", mock.Anything, loanID).Return(schedules, nil)
			},
			expectedError:      false,
			expectedDelinquent: false,
		},
		{
			name:   "Success - Not delinquent (consecutive missed but reset by payment)",
			loanID: "LOAN126",
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:        loanID,
					Amount:        decimal.NewFromInt(5000000),
					InterestRate:  decimal.NewFromFloat(0.10),
					DurationWeeks: 50,
					WeeklyPayment: decimal.NewFromInt(110000),
					Status:        domain.LoanStatusActive,
					CreatedAt:     time.Now().AddDate(0, 0, -28), // 4 weeks ago
				}

				schedules := []*domain.LoanSchedule{
					{ID: uuid.New(), LoanID: loanID, WeekNumber: 1, DueDate: time.Now().AddDate(0, 0, -21), DueAmount: decimal.NewFromInt(110000), Status: domain.ScheduleStatusPending},
					{ID: uuid.New(), LoanID: loanID, WeekNumber: 2, DueDate: time.Now().AddDate(0, 0, -14), DueAmount: decimal.NewFromInt(110000), Status: domain.ScheduleStatusPaid}, // Payment resets counter
					{ID: uuid.New(), LoanID: loanID, WeekNumber: 3, DueDate: time.Now().AddDate(0, 0, -7), DueAmount: decimal.NewFromInt(110000), Status: domain.ScheduleStatusPending},
					{ID: uuid.New(), LoanID: loanID, WeekNumber: 4, DueDate: time.Now(), DueAmount: decimal.NewFromInt(110000), Status: domain.ScheduleStatusPending}, // Current week
				}

				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
				mockLoanRepo.On("GetScheduleByLoanID", mock.Anything, loanID).Return(schedules, nil)
			},
			expectedError:      false,
			expectedDelinquent: false,
		},
		{
			name:   "Failure - Loan not found",
			loanID: "NONEXISTENT",
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(nil, sql.ErrNoRows)
			},
			expectedError:      true,
			errorContains:      "database",
			expectedDelinquent: false,
		},
		{
			name:   "Failure - Database error getting schedule",
			loanID: "LOAN127",
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:    loanID,
					Status:    domain.LoanStatusActive,
					CreatedAt: time.Now().AddDate(0, 0, -21),
				}

				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
				mockLoanRepo.On("GetScheduleByLoanID", mock.Anything, loanID).Return(nil, errors.New("schedule query failed"))
			},
			expectedError:      true,
			errorContains:      "database",
			expectedDelinquent: false,
		},
		{
			name:   "Failure - loan status is already closed",
			loanID: "LOAN127",
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:    loanID,
					Status:    domain.LoanStatusClosed,
					CreatedAt: time.Now().AddDate(0, 0, -21),
				}

				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
			},
			expectedError:      true,
			errorContains:      "closed",
			expectedDelinquent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockLoanRepo := &mocks.MockLoanRepository{}
			mockPaymentRepo := &mocks.MockPaymentRepository{}

			service := billingService.NewBillingService(mockLoanRepo, mockPaymentRepo, nil, nil)

			tt.setupMocks(mockLoanRepo, mockPaymentRepo, tt.loanID)

			// Act
			isDelinquent, err := service.IsDelinquent(context.Background(), tt.loanID)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDelinquent, isDelinquent)
			}

			mockLoanRepo.AssertExpectations(t)
			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestMakePayment(t *testing.T) {
	tests := []struct {
		name           string
		request        domain.MakePaymentRequest
		setupMocks     func(*mocks.MockLoanRepository, *mocks.MockPaymentRepository, string)
		expectedError  bool
		errorContains  string
		validateResult func(*testing.T, *domain.Payment)
	}{
		{
			name: "Success - First payment on active loan",
			request: domain.MakePaymentRequest{
				LoanID: "LOAN123",
				Amount: decimal.NewFromInt(110000),
			},
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:        loanID,
					Amount:        decimal.NewFromInt(5000000),
					InterestRate:  decimal.NewFromFloat(0.10),
					DurationWeeks: 50,
					WeeklyPayment: decimal.NewFromInt(110000),
					Status:        domain.LoanStatusActive,
				}

				schedules := []*domain.LoanSchedule{
					{LoanID: loanID, WeekNumber: 1, Status: domain.ScheduleStatusPending, DueAmount: decimal.NewFromInt(110000)},
					{LoanID: loanID, WeekNumber: 2, Status: domain.ScheduleStatusPending, DueAmount: decimal.NewFromInt(110000)},
				}

				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
				mockLoanRepo.On("GetScheduleByLoanID", mock.Anything, loanID).Return(schedules, nil)
				mockPaymentRepo.On("Create", mock.Anything, mock.MatchedBy(func(payment *domain.Payment) bool {
					return payment.LoanID == loanID && payment.Amount.Equal(decimal.NewFromInt(110000)) && payment.WeekNumber == 1
				})).Return(nil)
				mockLoanRepo.On("UpdateScheduleStatus", mock.Anything, loanID, 1, "PAID").Return(nil)
			},
			expectedError: false,
			validateResult: func(t *testing.T, payment *domain.Payment) {
				assert.Equal(t, "LOAN123", payment.LoanID)
				assert.True(t, payment.Amount.Equal(decimal.NewFromInt(110000)))
				assert.Equal(t, 1, payment.WeekNumber)
			},
		},
		{
			name: "Success - Final payment closes loan",
			request: domain.MakePaymentRequest{
				LoanID: "LOAN124",
				Amount: decimal.NewFromInt(110000),
			},
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:        loanID,
					Amount:        decimal.NewFromInt(5000000),
					InterestRate:  decimal.NewFromFloat(0.10),
					DurationWeeks: 2,
					WeeklyPayment: decimal.NewFromInt(110000),
					Status:        domain.LoanStatusActive,
				}

				schedules := []*domain.LoanSchedule{
					{LoanID: loanID, WeekNumber: 1, Status: "PAID", DueAmount: decimal.NewFromInt(110000)},
					{LoanID: loanID, WeekNumber: 2, Status: domain.ScheduleStatusPending, DueAmount: decimal.NewFromInt(110000)},
				}

				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
				mockLoanRepo.On("GetScheduleByLoanID", mock.Anything, loanID).Return(schedules, nil)
				mockPaymentRepo.On("Create", mock.Anything, mock.MatchedBy(func(payment *domain.Payment) bool {
					return payment.LoanID == loanID && payment.WeekNumber == 2
				})).Return(nil)
				mockLoanRepo.On("UpdateScheduleStatus", mock.Anything, loanID, 2, "PAID").Return(nil)
				mockLoanRepo.On("Update", mock.Anything, mock.MatchedBy(func(updatedLoan *domain.Loan) bool {
					return updatedLoan.Status == domain.LoanStatusClosed
				})).Return(nil)
			},
			expectedError: false,
			validateResult: func(t *testing.T, payment *domain.Payment) {
				assert.Equal(t, "LOAN124", payment.LoanID)
				assert.Equal(t, 2, payment.WeekNumber)
			},
		},
		{
			name: "Failure - Zero payment amount",
			request: domain.MakePaymentRequest{
				LoanID: "LOAN125",
				Amount: decimal.Zero,
			},
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				// No mocks needed - validation happens before database calls
			},
			expectedError: true,
			errorContains: "payment amount",
			validateResult: func(t *testing.T, payment *domain.Payment) {
				assert.Nil(t, payment)
			},
		},
		{
			name: "Failure - Negative payment amount",
			request: domain.MakePaymentRequest{
				LoanID: "LOAN126",
				Amount: decimal.NewFromInt(-50000),
			},
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				// No mocks needed - validation happens before database calls
			},
			expectedError: true,
			errorContains: "payment amount",
			validateResult: func(t *testing.T, payment *domain.Payment) {
				assert.Nil(t, payment)
			},
		},
		{
			name: "Failure - Loan already closed",
			request: domain.MakePaymentRequest{
				LoanID: "LOAN127",
				Amount: decimal.NewFromInt(110000),
			},
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:        loanID,
					Status:        domain.LoanStatusClosed,
					WeeklyPayment: decimal.NewFromInt(110000),
				}
				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
			},
			expectedError: true,
			errorContains: "closed",
			validateResult: func(t *testing.T, payment *domain.Payment) {
				assert.Nil(t, payment)
			},
		},
		{
			name: "Failure - Invalid payment amount (too high)",
			request: domain.MakePaymentRequest{
				LoanID: "LOAN128",
				Amount: decimal.NewFromInt(220000),
			},
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:        loanID,
					WeeklyPayment: decimal.NewFromInt(110000),
					Status:        domain.LoanStatusActive,
				}
				schedules := []*domain.LoanSchedule{
					{LoanID: loanID, WeekNumber: 1, Status: domain.ScheduleStatusPending, DueAmount: decimal.NewFromInt(110000)},
				}
				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
				mockLoanRepo.On("GetScheduleByLoanID", mock.Anything, loanID).Return(schedules, nil)
			},
			expectedError: true,
			errorContains: "payment amount",
			validateResult: func(t *testing.T, payment *domain.Payment) {
				assert.Nil(t, payment)
			},
		},
		{
			name: "Failure - No outstanding balance",
			request: domain.MakePaymentRequest{
				LoanID: "LOAN129",
				Amount: decimal.NewFromInt(110000),
			},
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				loan := &domain.Loan{
					LoanID:        loanID,
					WeeklyPayment: decimal.NewFromInt(110000),
					Status:        domain.LoanStatusActive,
				}
				schedules := []*domain.LoanSchedule{
					{LoanID: loanID, WeekNumber: 1, Status: "PAID", DueAmount: decimal.NewFromInt(110000)},
					{LoanID: loanID, WeekNumber: 2, Status: "PAID", DueAmount: decimal.NewFromInt(110000)},
				}
				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(loan, nil)
				mockLoanRepo.On("GetScheduleByLoanID", mock.Anything, loanID).Return(schedules, nil)
			},
			expectedError: true,
			errorContains: "outstanding balance",
			validateResult: func(t *testing.T, payment *domain.Payment) {
				assert.Nil(t, payment)
			},
		},
		{
			name: "Failure - Loan not found",
			request: domain.MakePaymentRequest{
				LoanID: "NONEXISTENT",
				Amount: decimal.NewFromInt(110000),
			},
			setupMocks: func(mockLoanRepo *mocks.MockLoanRepository, mockPaymentRepo *mocks.MockPaymentRepository, loanID string) {
				mockLoanRepo.On("GetByLoanID", mock.Anything, loanID).Return(nil, sql.ErrNoRows)
			},
			expectedError: true,
			validateResult: func(t *testing.T, payment *domain.Payment) {
				assert.Nil(t, payment)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockLoanRepo := &mocks.MockLoanRepository{}
			mockPaymentRepo := &mocks.MockPaymentRepository{}

			service := billingService.NewBillingService(mockLoanRepo, mockPaymentRepo, nil, nil)

			tt.setupMocks(mockLoanRepo, mockPaymentRepo, tt.request.LoanID)

			// Act
			payment, err := service.MakePayment(context.Background(), tt.request)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			tt.validateResult(t, payment)
			mockLoanRepo.AssertExpectations(t)
			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

package service

import (
	"context"
	"testing"

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

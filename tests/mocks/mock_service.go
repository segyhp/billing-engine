package mocks

import (
	"context"

	"github.com/segyhp/billing-engine/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
)

type MockBillingService struct {
	mock.Mock
}

func (m *MockBillingService) CreateLoan(ctx context.Context, request *domain.CreateLoanRequest) (*domain.Loan, []*domain.LoanSchedule, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*domain.Loan), args.Get(1).([]*domain.LoanSchedule), args.Error(2)
}

func (m *MockBillingService) GetOutstanding(ctx context.Context, loanID string) (decimal.Decimal, error) {
	args := m.Called(ctx, loanID)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockBillingService) IsDelinquent(ctx context.Context, loanID string) (bool, error) {
	args := m.Called(ctx, loanID)
	return args.Bool(0), args.Error(1)
}

func (m *MockBillingService) MakePayment(ctx context.Context, request domain.MakePaymentRequest) (*domain.Payment, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

// NewMockBillingService creates a new mock billing service instance
func NewMockBillingService() *MockBillingService {
	return &MockBillingService{}
}

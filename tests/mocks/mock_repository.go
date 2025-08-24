package mocks

import (
	"context"

	"github.com/segyhp/billing-engine/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockLoanRepository struct {
	mock.Mock
}

func (m *MockLoanRepository) Create(ctx context.Context, loan *domain.Loan) error {
	args := m.Called(ctx, loan)
	return args.Error(0)
}

func (m *MockLoanRepository) GetByLoanID(ctx context.Context, loanID string) (*domain.Loan, error) {
	args := m.Called(ctx, loanID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Loan), args.Error(1)
}

func (m *MockLoanRepository) CreateSchedule(ctx context.Context, schedules []*domain.LoanSchedule) error {
	args := m.Called(ctx, schedules)
	return args.Error(0)
}

func (m *MockLoanRepository) UpdateStatus(ctx context.Context, loanID string, status string) error {
	args := m.Called(ctx, loanID, status)
	return args.Error(0)
}

type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByLoanID(ctx context.Context, loanID string) ([]*domain.Payment, error) {
	args := m.Called(ctx, loanID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetByLoanIDAndWeek(ctx context.Context, loanID string, weekNumber int) (*domain.Payment, error) {
	args := m.Called(ctx, loanID, weekNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisClient) Del(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

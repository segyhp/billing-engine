package mocks

import (
	"context"
	"time"

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

func (m *MockLoanRepository) Update(ctx context.Context, loan *domain.Loan) error {
	args := m.Called(ctx, loan)
	return args.Error(0)
}

func (m *MockLoanRepository) CreateSchedule(ctx context.Context, schedules []*domain.LoanSchedule) error {
	args := m.Called(ctx, schedules)
	return args.Error(0)
}

func (m *MockLoanRepository) GetScheduleByLoanID(ctx context.Context, loanID string) ([]*domain.LoanSchedule, error) {
	args := m.Called(ctx, loanID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LoanSchedule), args.Error(1)
}

func (m *MockLoanRepository) UpdateScheduleStatus(ctx context.Context, loanID string, weekNumber int, status string) error {
	args := m.Called(ctx, loanID, weekNumber, status)
	return args.Error(0)
}

func (m *MockLoanRepository) GetOverdueSchedules(ctx context.Context, loanID string, currentDate time.Time) ([]*domain.LoanSchedule, error) {
	args := m.Called(ctx, loanID, currentDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.LoanSchedule), args.Error(1)
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

func (m *MockPaymentRepository) GetTotalPaid(ctx context.Context, loanID string) (float64, error) {
	args := m.Called(ctx, loanID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockPaymentRepository) GetLatestPayment(ctx context.Context, loanID string) (*domain.Payment, error) {
	args := m.Called(ctx, loanID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

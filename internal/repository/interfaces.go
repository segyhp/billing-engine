package repository

import (
	"context"
	"time"

	"github.com/segyhp/billing-engine/internal/domain"
)

// LoanRepository defines the interface for loan data operations
type LoanRepository interface {
	// Create creates a new loan
	Create(ctx context.Context, loan *domain.Loan) error

	// GetByLoanID retrieves a loan by its loan ID
	GetByLoanID(ctx context.Context, loanID string) (*domain.Loan, error)

	// Update updates a loan
	Update(ctx context.Context, loan *domain.Loan) error

	// CreateSchedule creates loan schedule entries
	CreateSchedule(ctx context.Context, schedules []*domain.LoanSchedule) error

	// GetScheduleByLoanID retrieves loan schedule by loan ID
	GetScheduleByLoanID(ctx context.Context, loanID string) ([]*domain.LoanSchedule, error)

	// UpdateScheduleStatus updates the status of a specific schedule entry
	UpdateScheduleStatus(ctx context.Context, loanID string, weekNumber int, status string) error

	// GetOverdueSchedules gets schedules that are overdue for a loan
	GetOverdueSchedules(ctx context.Context, loanID string, currentDate time.Time) ([]*domain.LoanSchedule, error)
}

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	// Create creates a new payment record
	Create(ctx context.Context, payment *domain.Payment) error

	// GetByLoanID retrieves all payments for a loan
	GetByLoanID(ctx context.Context, loanID string) ([]*domain.Payment, error)

	// GetTotalPaid calculates total amount paid for a loan
	GetTotalPaid(ctx context.Context, loanID string) (float64, error)

	// GetLatestPayment gets the most recent payment for a loan
	GetLatestPayment(ctx context.Context, loanID string) (*domain.Payment, error)
}

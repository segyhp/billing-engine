package repository

import (
	"context"
	"time"

	"github.com/segyhp/billing-engine/internal/domain"

	"github.com/jmoiron/sqlx"
)

type loanRepository struct {
	db *sqlx.DB
}

func NewLoanRepository(db *sqlx.DB) LoanRepository {
	return &loanRepository{db: db}
}

func (r *loanRepository) Create(ctx context.Context, loan *domain.Loan) error {
	//todo: use transaction
	return nil
}

func (r *loanRepository) GetByLoanID(ctx context.Context, loanID string) (*domain.Loan, error) {
	//todo: use transaction
	return nil, nil
}

func (r *loanRepository) Update(ctx context.Context, loan *domain.Loan) error {
	//todo: use transaction
	return nil
}

func (r *loanRepository) CreateSchedule(ctx context.Context, schedules []*domain.LoanSchedule) error {
	//todo: use transaction
	return nil
}

func (r *loanRepository) GetScheduleByLoanID(ctx context.Context, loanID string) ([]*domain.LoanSchedule, error) {
	//todo: implement
	return nil, nil
}

func (r *loanRepository) UpdateScheduleStatus(ctx context.Context, loanID string, weekNumber int, status string) error {
	//todo: use transaction
	return nil
}

func (r *loanRepository) GetOverdueSchedules(ctx context.Context, loanID string, currentDate time.Time) ([]*domain.LoanSchedule, error) {
	//todo: implement
	return nil, nil
}

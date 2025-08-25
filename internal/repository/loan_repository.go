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
	query := `
		INSERT INTO loans (id, loan_id, amount, interest_rate, duration_weeks, weekly_payment, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		loan.ID,
		loan.LoanID,
		loan.Amount,
		loan.InterestRate,
		loan.DurationWeeks,
		loan.WeeklyPayment,
		loan.Status,
		loan.CreatedAt,
		loan.UpdatedAt,
	)

	return err
}

func (r *loanRepository) GetByLoanID(ctx context.Context, loanID string) (*domain.Loan, error) {
	query := `
		SELECT id, loan_id, amount, interest_rate, duration_weeks, weekly_payment, status, created_at, updated_at
		FROM loans
		WHERE loan_id = $1
	`

	var loan domain.Loan
	err := r.db.GetContext(ctx, &loan, query, loanID)
	if err != nil {
		return nil, err
	}

	return &loan, nil
}

func (r *loanRepository) Update(ctx context.Context, loan *domain.Loan) error {
	query := `
		UPDATE loans
		SET amount = $2, interest_rate = $3, duration_weeks = $4, weekly_payment = $5, status = $6, updated_at = $7
		WHERE loan_id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		loan.LoanID,
		loan.Amount,
		loan.InterestRate,
		loan.DurationWeeks,
		loan.WeeklyPayment,
		loan.Status,
		time.Now(),
	)

	return err
}

func (r *loanRepository) CreateSchedule(ctx context.Context, schedules []*domain.LoanSchedule) error {
	query := `
		INSERT INTO loan_schedule (id, loan_id, week_number, due_amount, due_date, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, schedule := range schedules {
		_, err = tx.ExecContext(ctx, query,
			schedule.ID,
			schedule.LoanID,
			schedule.WeekNumber,
			schedule.DueAmount,
			schedule.DueDate,
			schedule.Status,
			schedule.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *loanRepository) GetScheduleByLoanID(ctx context.Context, loanID string) ([]*domain.LoanSchedule, error) {
	query := `
		SELECT id, loan_id, week_number, due_amount, due_date, status, created_at
		FROM loan_schedule
		WHERE loan_id = $1
		ORDER BY week_number
	`

	var schedules []*domain.LoanSchedule
	err := r.db.SelectContext(ctx, &schedules, query, loanID)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

func (r *loanRepository) UpdateScheduleStatus(ctx context.Context, loanID string, weekNumber int, status string) error {
	query := `
		UPDATE loan_schedule
		SET status = $3
		WHERE loan_id = $1 AND week_number = $2
	`

	_, err := r.db.ExecContext(ctx, query, loanID, weekNumber, status)
	return err
}

func (r *loanRepository) GetOverdueSchedules(ctx context.Context, loanID string, currentDate time.Time) ([]*domain.LoanSchedule, error) {
	query := `
		SELECT id, loan_id, week_number, due_amount, due_date, status, created_at
		FROM loan_schedule
		WHERE loan_id = $1 AND status = 'pending' AND due_date < $2
		ORDER BY week_number
	`

	var schedules []*domain.LoanSchedule
	err := r.db.SelectContext(ctx, &schedules, query, loanID, currentDate)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

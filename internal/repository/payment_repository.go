package repository

import (
	"context"

	"github.com/segyhp/billing-engine/internal/domain"

	"github.com/jmoiron/sqlx"
)

type paymentRepository struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	query := `
		INSERT INTO payments (id, loan_id, amount, payment_date, week_number, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		payment.ID,
		payment.LoanID,
		payment.Amount,
		payment.PaymentDate,
		payment.WeekNumber,
		payment.CreatedAt,
	)

	return err
}

func (r *paymentRepository) GetByLoanID(ctx context.Context, loanID string) ([]*domain.Payment, error) {
	query := `
		SELECT id, loan_id, amount, payment_date, week_number, created_at
		FROM payments
		WHERE loan_id = $1
		ORDER BY payment_date DESC
	`

	var payments []*domain.Payment
	err := r.db.SelectContext(ctx, &payments, query, loanID)
	if err != nil {
		return nil, err
	}

	return payments, nil
}

func (r *paymentRepository) GetTotalPaid(ctx context.Context, loanID string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) as total_paid
		FROM payments
		WHERE loan_id = $1
	`

	var totalPaid float64
	err := r.db.GetContext(ctx, &totalPaid, query, loanID)
	if err != nil {
		return 0, err
	}

	return totalPaid, nil
}

func (r *paymentRepository) GetLatestPayment(ctx context.Context, loanID string) (*domain.Payment, error) {
	query := `
		SELECT id, loan_id, amount, payment_date, week_number, created_at
		FROM payments
		WHERE loan_id = $1
		ORDER BY payment_date DESC, created_at DESC
		LIMIT 1
	`

	var payment domain.Payment
	err := r.db.GetContext(ctx, &payment, query, loanID)
	if err != nil {
		return nil, err
	}

	return &payment, nil
}

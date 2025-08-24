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
	//todo: use transaction
	return nil
}

func (r *paymentRepository) GetByLoanID(ctx context.Context, loanID string) ([]*domain.Payment, error) {
	//todo: implement
	return nil, nil
}

func (r *paymentRepository) GetTotalPaid(ctx context.Context, loanID string) (float64, error) {
	//todo: implement
	return 0, nil
}

func (r *paymentRepository) GetLatestPayment(ctx context.Context, loanID string) (*domain.Payment, error) {
	//todo: implement
	return nil, nil
}

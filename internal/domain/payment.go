package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Payment represents a payment made by a borrower
type Payment struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	LoanID      string          `json:"loan_id" db:"loan_id"`
	Amount      decimal.Decimal `json:"amount" db:"amount"`
	PaymentDate time.Time       `json:"payment_date" db:"payment_date"`
	WeekNumber  int             `json:"week_number" db:"week_number"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

type MakePaymentRequest struct {
	LoanID string          `json:"loan_id" validate:"required"`
	Amount decimal.Decimal `json:"amount" validate:"required,gt=0"`
}

type MakePaymentResponse struct {
	Payment        *Payment        `json:"payment"`
	Outstanding    decimal.Decimal `json:"outstanding"`
	IsDelinquent   bool            `json:"is_delinquent"`
	PaidWeekNumber int             `json:"paid_week_number"`
}

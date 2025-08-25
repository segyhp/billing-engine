package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	LoanStatusActive  = "active"
	LoanStatusClosed  = "closed"
	LoanStatusDefault = "default"
)

// Loan represents a loan entity
type Loan struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	LoanID        string          `json:"loan_id" db:"loan_id"`
	Amount        decimal.Decimal `json:"amount" db:"amount"`
	InterestRate  decimal.Decimal `json:"interest_rate" db:"interest_rate"`
	DurationWeeks int             `json:"duration_weeks" db:"duration_weeks"`
	WeeklyPayment decimal.Decimal `json:"weekly_payment" db:"weekly_payment"`
	Status        string          `json:"status" db:"status"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`
}

// DTOs for requests and responses

type CreateLoanRequest struct {
	LoanID        string          `json:"loan_id" validate:"required"`
	Amount        decimal.Decimal `json:"amount" validate:"required,decimal_gt=0"`
	InterestRate  decimal.Decimal `json:"interest_rate" validate:"required,decimal_gte=0"`
	DurationWeeks int             `json:"duration_weeks" validate:"required,gt=0"`
}

type CreateLoanResponse struct {
	Loan     *Loan           `json:"loan"`
	Schedule []*LoanSchedule `json:"schedule"`
}

type OutstandingResponse struct {
	LoanID      string          `json:"loan_id"`
	Outstanding decimal.Decimal `json:"outstanding"`
}

type DelinquentResponse struct {
	LoanID       string `json:"loan_id"`
	IsDelinquent bool   `json:"is_delinquent"`
	MissedWeeks  int    `json:"missed_weeks"`
}

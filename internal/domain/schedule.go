package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Business logic constants
const (
	ScheduleStatusPending = "pending"
	ScheduleStatusPaid    = "paid"
	ScheduleStatusOverdue = "overdue"
)

// LoanSchedule represents a loan schedule entry
type LoanSchedule struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	LoanID     string          `json:"loan_id" db:"loan_id"`
	WeekNumber int             `json:"week_number" db:"week_number"`
	DueAmount  decimal.Decimal `json:"due_amount" db:"due_amount"`
	DueDate    time.Time       `json:"due_date" db:"due_date"`
	Status     string          `json:"status" db:"status"` // pending, paid, overdue
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

type ScheduleResponse struct {
	LoanID   string          `json:"loan_id"`
	Schedule []*LoanSchedule `json:"schedule"`
}

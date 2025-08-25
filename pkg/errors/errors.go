package errors

import (
	"errors"
	"fmt"
)

// Domain errors
var (
	ErrLoanNotFound          = errors.New("loan not found")
	ErrLoanAlreadyExists     = errors.New("loan already exists")
	ErrInvalidLoanAmount     = errors.New("invalid loan amount")
	ErrInvalidPaymentAmount  = errors.New("invalid payment amount")
	ErrLoanAlreadyClosed     = errors.New("loan is already closed")
	ErrPaymentAmountMismatch = errors.New("payment amount must match weekly payment amount exactly")
	ErrNoOutstandingBalance  = errors.New("no outstanding balance")
)

// BusinessError represents a business logic error
type BusinessError struct {
	Code    string
	Message string
	Err     error
}

func (e *BusinessError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *BusinessError) Unwrap() error {
	return e.Err
}

// NewBusinessError creates a new business error
func NewBusinessError(code, message string, err error) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Error codes
const (
	ErrCodeLoanNotFound          = "LOAN_NOT_FOUND"
	ErrCodeLoanAlreadyExists     = "LOAN_ALREADY_EXISTS"
	ErrCodeInvalidLoanAmount     = "INVALID_LOAN_AMOUNT"
	ErrCodeInvalidPaymentAmount  = "INVALID_PAYMENT_AMOUNT"
	ErrCodeLoanAlreadyClosed     = "LOAN_ALREADY_CLOSED"
	ErrCodePaymentAmountMismatch = "PAYMENT_AMOUNT_MISMATCH"
	ErrCodeNoOutstandingBalance  = "NO_OUTSTANDING_BALANCE"
	ErrCodeDatabaseError         = "DATABASE_ERROR"
	ErrCodeCacheError            = "CACHE_ERROR"
)

// Wrap common errors with business context
func WrapLoanNotFound(loanID string) *BusinessError {
	return NewBusinessError(
		ErrCodeLoanNotFound,
		fmt.Sprintf("Loan with ID %s not found", loanID),
		ErrLoanNotFound,
	)
}

func WrapLoanAlreadyExists(loanID string) *BusinessError {
	return NewBusinessError(
		ErrCodeLoanAlreadyExists,
		fmt.Sprintf("Loan with ID %s already exists", loanID),
		ErrLoanAlreadyExists,
	)
}

func WrapPaymentAmountMismatch(expected, actual string) *BusinessError {
	return NewBusinessError(
		ErrCodePaymentAmountMismatch,
		fmt.Sprintf("Payment amount %s does not match expected weekly payment %s", actual, expected),
		ErrPaymentAmountMismatch,
	)
}

func WrapLoanAlreadyClosed(loanID string) *BusinessError {
	return NewBusinessError(
		ErrCodeLoanAlreadyClosed,
		fmt.Sprintf("Loan with ID %s is already closed", loanID),
		ErrLoanAlreadyClosed,
	)
}

func WrapDatabaseError(err error) *BusinessError {
	return NewBusinessError(
		ErrCodeDatabaseError,
		"database operation failed",
		err,
	)
}

func WrapCacheError(err error) *BusinessError {
	return NewBusinessError(
		ErrCodeCacheError,
		"Cache operation failed",
		err,
	)
}

func WrapNoOutstandingBalance(loanID string) *BusinessError {
	return NewBusinessError(
		ErrCodeNoOutstandingBalance,
		fmt.Sprintf("Loan with ID %s has no outstanding balance", loanID),
		ErrNoOutstandingBalance,
	)
}

func WrapInvalidPaymentAmount(amount float64) *BusinessError {
	return NewBusinessError(
		ErrCodeInvalidPaymentAmount,
		fmt.Sprintf("Invalid payment amount: %.2f", amount),
		ErrInvalidPaymentAmount,
	)
}

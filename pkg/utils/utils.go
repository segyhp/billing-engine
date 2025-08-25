package utils

import (
	"time"

	"github.com/shopspring/decimal"
)

// CalculateWeeklyPayment calculates the weekly payment amount
// Formula: (Principal + Interest) / Duration
func CalculateWeeklyPayment(principal decimal.Decimal, annualRate decimal.Decimal, weeks int) decimal.Decimal {
	totalInterest := principal.Mul(annualRate)
	totalAmount := principal.Add(totalInterest)
	weeklyPayment := totalAmount.Div(decimal.NewFromInt(int64(weeks)))

	// Round to 2 decimal places
	return weeklyPayment.Round(2)
}

// CalculateDueDate calculates the due date for a specific week
// Assumes weekly payments are due every 7 days starting from loan creation
func CalculateDueDate(loanStartDate time.Time, weekNumber int) time.Time {
	days := weekNumber * 7 // Week 1 is due 7 days after start, Week 2 is due 14 days after, etc.
	return loanStartDate.AddDate(0, 0, days)
}

// GetCurrentWeek calculates which week we're currently in based on loan start date
func GetCurrentWeek(loanStartDate time.Time, loanEndDate time.Time) int {
	duration := loanEndDate.Sub(loanStartDate)
	days := int(duration.Hours() / 24)
	week := (days / 7) + 1

	if week < 1 {
		return 1
	}

	return week
}

// IsDateOverdue checks if a date is overdue (past current date)
func IsDateOverdue(dueDate time.Time) bool {
	return time.Now().After(dueDate)
}

// DecimalFromFloat converts float64 to decimal.Decimal
func DecimalFromFloat(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

// DecimalFromString converts string to decimal.Decimal
func DecimalFromString(s string) (decimal.Decimal, error) {
	return decimal.NewFromString(s)
}

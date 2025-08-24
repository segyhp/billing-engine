package utils

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalculateWeeklyPayment(t *testing.T) {
	tests := []struct {
		name      string
		principal decimal.Decimal
		rate      decimal.Decimal
		weeks     int
		expected  decimal.Decimal
	}{
		{
			name:      "standard loan calculation",
			principal: decimal.NewFromInt(5000000),
			rate:      decimal.NewFromFloat(0.10),
			weeks:     50,
			expected:  decimal.NewFromInt(110000), // (5,000,000 * 1.10) / 50 = 110,000
		},
		{
			name:      "smaller loan",
			principal: decimal.NewFromInt(1000000),
			rate:      decimal.NewFromFloat(0.10),
			weeks:     10,
			expected:  decimal.NewFromInt(110000), // (1,000,000 * 1.10) / 10 = 110,000
		},
		{
			name:      "zero interest rate",
			principal: decimal.NewFromInt(5000000),
			rate:      decimal.NewFromInt(0),
			weeks:     50,
			expected:  decimal.NewFromInt(100000), // 5,000,000 / 50 = 100,000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test will fail initially (RED) - implement the function to make it pass (GREEN)
			result := CalculateWeeklyPayment(tt.principal, tt.rate, tt.weeks)
			assert.True(t, result.Equal(tt.expected),
				"Expected %v, but got %v", tt.expected, result)
		})
	}
}

func TestCalculateDueDate(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		startDate  time.Time
		weekNumber int
		expected   time.Time
	}{
		{
			name:       "first week",
			startDate:  baseDate,
			weekNumber: 1,
			expected:   baseDate.AddDate(0, 0, 7), // 7 days later
		},
		{
			name:       "second week",
			startDate:  baseDate,
			weekNumber: 2,
			expected:   baseDate.AddDate(0, 0, 14), // 14 days later
		},
		{
			name:       "week 50",
			startDate:  baseDate,
			weekNumber: 50,
			expected:   baseDate.AddDate(0, 0, 350), // 350 days later
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateDueDate(tt.startDate, tt.weekNumber)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCurrentWeek(t *testing.T) {
	loanStartDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		startDate   time.Time
		currentDate time.Time
		expected    int
	}{
		{
			name:        "same day as loan start",
			startDate:   loanStartDate,
			currentDate: loanStartDate,
			expected:    1,
		},
		{
			name:        "one week later",
			startDate:   loanStartDate,
			currentDate: loanStartDate.AddDate(0, 0, 7),
			expected:    2,
		},
		{
			name:        "middle of second week",
			startDate:   loanStartDate,
			currentDate: loanStartDate.AddDate(0, 0, 10),
			expected:    2,
		},
		{
			name:        "week 50",
			startDate:   loanStartDate,
			currentDate: loanStartDate.AddDate(0, 0, 343), // 49 * 7 = 343
			expected:    50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCurrentWeek(tt.startDate, tt.currentDate)
			assert.Equal(t, tt.expected, result)
		})
	}
}

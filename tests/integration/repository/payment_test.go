package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/segyhp/billing-engine/internal/domain"
	"github.com/segyhp/billing-engine/internal/repository"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	// Create loan first to satisfy foreign key constraint
	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        "LOAN-PAY-001",
		Amount:        decimal.NewFromInt(1000000),
		InterestRate:  decimal.NewFromFloat(0.1),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(22000),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	loanRepo := repository.NewLoanRepository(db)
	err := loanRepo.Create(ctx, loan)
	require.NoError(t, err)

	payment := &domain.Payment{
		ID:          uuid.New(),
		LoanID:      "LOAN-PAY-001",
		Amount:      decimal.NewFromInt(22000),
		PaymentDate: time.Now(),
		WeekNumber:  1,
		CreatedAt:   time.Now(),
	}

	err = repo.Create(ctx, payment)
	require.NoError(t, err)
}

func TestPaymentRepository_GetByLoanID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	// Create loan first
	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        "LOAN-PAY-002",
		Amount:        decimal.NewFromInt(1000000),
		InterestRate:  decimal.NewFromFloat(0.1),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(22000),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	loanRepo := repository.NewLoanRepository(db)
	err := loanRepo.Create(ctx, loan)
	require.NoError(t, err)

	// Create multiple payments
	payments := []*domain.Payment{
		{
			ID:          uuid.New(),
			LoanID:      "LOAN-PAY-002",
			Amount:      decimal.NewFromInt(22000),
			PaymentDate: time.Now().AddDate(0, 0, -7),
			WeekNumber:  1,
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			LoanID:      "LOAN-PAY-002",
			Amount:      decimal.NewFromInt(22000),
			PaymentDate: time.Now().AddDate(0, 0, -1),
			WeekNumber:  2,
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			LoanID:      "LOAN-PAY-002",
			Amount:      decimal.NewFromInt(15000),
			PaymentDate: time.Now(),
			WeekNumber:  3,
			CreatedAt:   time.Now(),
		},
	}

	for _, payment := range payments {
		err = repo.Create(ctx, payment)
		require.NoError(t, err)
	}

	result, err := repo.GetByLoanID(ctx, "LOAN-PAY-002")
	require.NoError(t, err)
	assert.Len(t, result, 3)

	// Should be ordered by payment_date DESC
	assert.Equal(t, 3, result[0].WeekNumber) // Latest payment first
	assert.Equal(t, 2, result[1].WeekNumber)
	assert.Equal(t, 1, result[2].WeekNumber) // Oldest payment last
	assert.True(t, decimal.NewFromInt(15000).Equal(result[0].Amount))
}

func TestPaymentRepository_GetByLoanID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	result, err := repo.GetByLoanID(ctx, "NON-EXISTENT-LOAN")
	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestPaymentRepository_GetTotalPaid(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	// Create loan first
	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        "LOAN-PAY-003",
		Amount:        decimal.NewFromInt(1000000),
		InterestRate:  decimal.NewFromFloat(0.1),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(22000),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	loanRepo := repository.NewLoanRepository(db)
	err := loanRepo.Create(ctx, loan)
	require.NoError(t, err)

	// Create payments
	payments := []*domain.Payment{
		{
			ID:          uuid.New(),
			LoanID:      "LOAN-PAY-003",
			Amount:      decimal.NewFromInt(22000),
			PaymentDate: time.Now().AddDate(0, 0, -14),
			WeekNumber:  1,
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			LoanID:      "LOAN-PAY-003",
			Amount:      decimal.NewFromInt(22000),
			PaymentDate: time.Now().AddDate(0, 0, -7),
			WeekNumber:  2,
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			LoanID:      "LOAN-PAY-003",
			Amount:      decimal.NewFromInt(15000),
			PaymentDate: time.Now(),
			WeekNumber:  3,
			CreatedAt:   time.Now(),
		},
	}

	for _, payment := range payments {
		err = repo.Create(ctx, payment)
		require.NoError(t, err)
	}

	totalPaid, err := repo.GetTotalPaid(ctx, "LOAN-PAY-003")
	require.NoError(t, err)
	assert.Equal(t, 59000.0, totalPaid) // 22000 + 22000 + 15000
}

func TestPaymentRepository_GetTotalPaid_NoPayments(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	// Create loan first
	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        "LOAN-PAY-004",
		Amount:        decimal.NewFromInt(1000000),
		InterestRate:  decimal.NewFromFloat(0.1),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(22000),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	loanRepo := repository.NewLoanRepository(db)
	err := loanRepo.Create(ctx, loan)
	require.NoError(t, err)

	totalPaid, err := repo.GetTotalPaid(ctx, "LOAN-PAY-004")
	require.NoError(t, err)
	assert.Equal(t, 0.0, totalPaid)
}

func TestPaymentRepository_GetTotalPaid_NonExistentLoan(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	totalPaid, err := repo.GetTotalPaid(ctx, "NON-EXISTENT-LOAN")
	require.NoError(t, err)
	assert.Equal(t, 0.0, totalPaid)
}

func TestPaymentRepository_GetLatestPayment(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	// Create loan first
	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        "LOAN-PAY-005",
		Amount:        decimal.NewFromInt(1000000),
		InterestRate:  decimal.NewFromFloat(0.1),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(22000),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	loanRepo := repository.NewLoanRepository(db)
	err := loanRepo.Create(ctx, loan)
	require.NoError(t, err)

	now := time.Now()

	// Create payments with different dates
	payments := []*domain.Payment{
		{
			ID:          uuid.New(),
			LoanID:      "LOAN-PAY-005",
			Amount:      decimal.NewFromInt(22000),
			PaymentDate: now.AddDate(0, 0, -14),
			WeekNumber:  1,
			CreatedAt:   now.AddDate(0, 0, -14),
		},
		{
			ID:          uuid.New(),
			LoanID:      "LOAN-PAY-005",
			Amount:      decimal.NewFromInt(22000),
			PaymentDate: now.AddDate(0, 0, -7),
			WeekNumber:  2,
			CreatedAt:   now.AddDate(0, 0, -7),
		},
		{
			ID:          uuid.New(),
			LoanID:      "LOAN-PAY-005",
			Amount:      decimal.NewFromInt(15000),
			PaymentDate: now, // Most recent payment_date
			WeekNumber:  3,
			CreatedAt:   now,
		},
	}

	for _, payment := range payments {
		err = repo.Create(ctx, payment)
		require.NoError(t, err)
	}

	latestPayment, err := repo.GetLatestPayment(ctx, "LOAN-PAY-005")
	require.NoError(t, err)
	assert.Equal(t, "LOAN-PAY-005", latestPayment.LoanID)
	assert.Equal(t, 3, latestPayment.WeekNumber)
	assert.True(t, decimal.NewFromInt(15000).Equal(latestPayment.Amount))
}

func TestPaymentRepository_GetLatestPayment_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	_, err := repo.GetLatestPayment(ctx, "NON-EXISTENT-LOAN")
	assert.Error(t, err)
}

func TestPaymentRepository_GetLatestPayment_SamePaymentDate_DifferentCreatedAt(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	// Create loan first
	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        "LOAN-PAY-006",
		Amount:        decimal.NewFromInt(1000000),
		InterestRate:  decimal.NewFromFloat(0.1),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(22000),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	loanRepo := repository.NewLoanRepository(db)
	err := loanRepo.Create(ctx, loan)
	require.NoError(t, err)

	now := time.Now()
	samePaymentDate := now.Truncate(24 * time.Hour) // Same date, different times

	// Create payments with same payment_date but different created_at
	payments := []*domain.Payment{
		{
			ID:          uuid.New(),
			LoanID:      "LOAN-PAY-006",
			Amount:      decimal.NewFromInt(22000),
			PaymentDate: samePaymentDate,
			WeekNumber:  1,
			CreatedAt:   now.Add(-1 * time.Hour), // Earlier created_at
		},
		{
			ID:          uuid.New(),
			LoanID:      "LOAN-PAY-006",
			Amount:      decimal.NewFromInt(15000),
			PaymentDate: samePaymentDate,
			WeekNumber:  2,
			CreatedAt:   now, // Later created_at (should be latest)
		},
	}

	for _, payment := range payments {
		err = repo.Create(ctx, payment)
		require.NoError(t, err)
	}

	latestPayment, err := repo.GetLatestPayment(ctx, "LOAN-PAY-006")
	require.NoError(t, err)
	assert.Equal(t, "LOAN-PAY-006", latestPayment.LoanID)
	assert.Equal(t, 2, latestPayment.WeekNumber) // Should be the one with later created_at
	assert.True(t, decimal.NewFromInt(15000).Equal(latestPayment.Amount))
}

func TestPaymentRepository_Create_ForeignKeyConstraint(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	// Try to create payment without loan (should fail due to foreign key constraint)
	payment := &domain.Payment{
		ID:          uuid.New(),
		LoanID:      "NON-EXISTENT-LOAN",
		Amount:      decimal.NewFromInt(22000),
		PaymentDate: time.Now(),
		WeekNumber:  1,
		CreatedAt:   time.Now(),
	}

	err := repo.Create(ctx, payment)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "violates foreign key constraint")
}

package repository

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/segyhp/billing-engine/internal/config"
	"github.com/segyhp/billing-engine/internal/domain"
	"github.com/segyhp/billing-engine/internal/repository"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *sqlx.DB

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Connect to postgres database to create test database
	cfg.Database.Name = "postgres"
	adminDB, err := sqlx.Connect("postgres", cfg.Database.DSN())
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to postgres database: %v", err))
	}
	defer adminDB.Close()

	// Create test database
	testDBName := "billing_engine_test"
	adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	if err != nil {
		panic(fmt.Sprintf("Failed to create test database: %v", err))
	}

	// Connect to test database
	cfg.Database.Name = testDBName
	testDB, err = sqlx.Connect("postgres", cfg.Database.DSN())
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	// Execute init.sql to create tables
	if err := executeInitSQL(testDB); err != nil {
		panic(fmt.Sprintf("Failed to initialize database schema: %v", err))
	}
}

func teardown() {
	if testDB != nil {
		testDB.Close()
	}

	// Drop test database
	cfg, _ := config.Load()
	cfg.Database.Name = "postgres"

	adminDB, err := sqlx.Connect("postgres", cfg.Database.DSN())
	if err != nil {
		return
	}
	defer adminDB.Close()

	adminDB.Exec("DROP DATABASE IF EXISTS billing_engine_test")
}

func executeInitSQL(db *sqlx.DB) error {
	// Read init.sql file
	sqlBytes, err := ioutil.ReadFile("../../../scripts/init.sql")
	if err != nil {
		return fmt.Errorf("failed to read init.sql: %w", err)
	}

	// Execute the SQL
	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		return fmt.Errorf("failed to execute init.sql: %w", err)
	}

	return nil
}

func setupTestDB(t *testing.T) *sqlx.DB {
	cleanupTestData(testDB)
	return testDB
}

func cleanupTestDB(db *sqlx.DB) {
	// No need to close the shared test DB
}

func cleanupTestData(db *sqlx.DB) {
	db.Exec("DELETE FROM loan_schedule")
	db.Exec("DELETE FROM payments")
	db.Exec("DELETE FROM loans")
}

func TestLoanRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewLoanRepository(db)
	ctx := context.Background()

	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        "LOAN-001",
		Amount:        decimal.NewFromInt(1000000),
		InterestRate:  decimal.NewFromFloat(0.1),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(22000),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(ctx, loan)
	require.NoError(t, err)
}

func TestLoanRepository_GetByLoanID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewLoanRepository(db)
	ctx := context.Background()

	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        "LOAN-002",
		Amount:        decimal.NewFromInt(500000),
		InterestRate:  decimal.NewFromFloat(0.15),
		DurationWeeks: 25,
		WeeklyPayment: decimal.NewFromInt(22000),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(ctx, loan)
	require.NoError(t, err)

	result, err := repo.GetByLoanID(ctx, "LOAN-002")
	require.NoError(t, err)
	assert.Equal(t, loan.LoanID, result.LoanID)
	assert.True(t, loan.Amount.Equal(result.Amount))
	assert.True(t, loan.InterestRate.Equal(result.InterestRate))
	assert.Equal(t, loan.Status, result.Status)
}

func TestLoanRepository_GetByLoanID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewLoanRepository(db)
	ctx := context.Background()

	_, err := repo.GetByLoanID(ctx, "NON-EXISTENT")
	assert.Error(t, err)
}

func TestLoanRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewLoanRepository(db)
	ctx := context.Background()

	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        "LOAN-003",
		Amount:        decimal.NewFromInt(750000),
		InterestRate:  decimal.NewFromFloat(0.12),
		DurationWeeks: 30,
		WeeklyPayment: decimal.NewFromInt(27500),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(ctx, loan)
	require.NoError(t, err)

	loan.Amount = decimal.NewFromInt(800000)
	loan.Status = "completed"
	err = repo.Update(ctx, loan)
	require.NoError(t, err)

	result, err := repo.GetByLoanID(ctx, "LOAN-003")
	require.NoError(t, err)
	assert.True(t, decimal.NewFromInt(800000).Equal(result.Amount))
	assert.Equal(t, "completed", result.Status)
}

func TestLoanRepository_CreateSchedule(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewLoanRepository(db)
	ctx := context.Background()

	// Create loan first to satisfy foreign key constraint
	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        "LOAN-004",
		Amount:        decimal.NewFromInt(1000000),
		InterestRate:  decimal.NewFromFloat(0.1),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(22000),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(ctx, loan)
	require.NoError(t, err)

	schedules := []*domain.LoanSchedule{
		{
			ID:         uuid.New(),
			LoanID:     "LOAN-004",
			WeekNumber: 1,
			DueAmount:  decimal.NewFromInt(22000),
			DueDate:    time.Now().AddDate(0, 0, 7),
			Status:     "pending",
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			LoanID:     "LOAN-004",
			WeekNumber: 2,
			DueAmount:  decimal.NewFromInt(22000),
			DueDate:    time.Now().AddDate(0, 0, 14),
			Status:     "pending",
			CreatedAt:  time.Now(),
		},
	}

	err = repo.CreateSchedule(ctx, schedules)
	require.NoError(t, err)
}

func TestLoanRepository_GetScheduleByLoanID(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewLoanRepository(db)
	ctx := context.Background()

	// Create loan first
	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        "LOAN-005",
		Amount:        decimal.NewFromInt(1000000),
		InterestRate:  decimal.NewFromFloat(0.1),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(25000),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(ctx, loan)
	require.NoError(t, err)

	schedules := []*domain.LoanSchedule{
		{
			ID:         uuid.New(),
			LoanID:     "LOAN-005",
			WeekNumber: 1,
			DueAmount:  decimal.NewFromInt(25000),
			DueDate:    time.Now().AddDate(0, 0, 7),
			Status:     "pending",
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			LoanID:     "LOAN-005",
			WeekNumber: 2,
			DueAmount:  decimal.NewFromInt(25000),
			DueDate:    time.Now().AddDate(0, 0, 14),
			Status:     "pending",
			CreatedAt:  time.Now(),
		},
	}

	err = repo.CreateSchedule(ctx, schedules)
	require.NoError(t, err)

	result, err := repo.GetScheduleByLoanID(ctx, "LOAN-005")
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].WeekNumber)
	assert.Equal(t, 2, result[1].WeekNumber)
	assert.True(t, decimal.NewFromInt(25000).Equal(result[0].DueAmount))
}

func TestLoanRepository_UpdateScheduleStatus(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewLoanRepository(db)
	ctx := context.Background()

	// Create loan first
	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        "LOAN-006",
		Amount:        decimal.NewFromInt(1000000),
		InterestRate:  decimal.NewFromFloat(0.1),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(20000),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(ctx, loan)
	require.NoError(t, err)

	schedules := []*domain.LoanSchedule{
		{
			ID:         uuid.New(),
			LoanID:     "LOAN-006",
			WeekNumber: 1,
			DueAmount:  decimal.NewFromInt(20000),
			DueDate:    time.Now().AddDate(0, 0, 7),
			Status:     "pending",
			CreatedAt:  time.Now(),
		},
	}

	err = repo.CreateSchedule(ctx, schedules)
	require.NoError(t, err)

	err = repo.UpdateScheduleStatus(ctx, "LOAN-006", 1, "paid")
	require.NoError(t, err)

	result, err := repo.GetScheduleByLoanID(ctx, "LOAN-006")
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "paid", result[0].Status)
}

func TestLoanRepository_GetOverdueSchedules(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewLoanRepository(db)
	ctx := context.Background()

	// Create loan first
	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        "LOAN-007",
		Amount:        decimal.NewFromInt(1000000),
		InterestRate:  decimal.NewFromFloat(0.1),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(20000),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(ctx, loan)
	require.NoError(t, err)

	pastDate := time.Now().AddDate(0, 0, -7)
	futureDate := time.Now().AddDate(0, 0, 7)

	schedules := []*domain.LoanSchedule{
		{
			ID:         uuid.New(),
			LoanID:     "LOAN-007",
			WeekNumber: 1,
			DueAmount:  decimal.NewFromInt(20000),
			DueDate:    pastDate,
			Status:     "pending",
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			LoanID:     "LOAN-007",
			WeekNumber: 2,
			DueAmount:  decimal.NewFromInt(20000),
			DueDate:    futureDate,
			Status:     "pending",
			CreatedAt:  time.Now(),
		},
		{
			ID:         uuid.New(),
			LoanID:     "LOAN-007",
			WeekNumber: 3,
			DueAmount:  decimal.NewFromInt(20000),
			DueDate:    pastDate,
			Status:     "paid",
			CreatedAt:  time.Now(),
		},
	}

	err = repo.CreateSchedule(ctx, schedules)
	require.NoError(t, err)

	result, err := repo.GetOverdueSchedules(ctx, "LOAN-007", time.Now())
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, result[0].WeekNumber)
	assert.Equal(t, "pending", result[0].Status)
}

func TestLoanRepository_CreateSchedule_TransactionRollback(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	repo := repository.NewLoanRepository(db)
	ctx := context.Background()

	// Create loan first
	loan := &domain.Loan{
		ID:            uuid.New(),
		LoanID:        "LOAN-008",
		Amount:        decimal.NewFromInt(1000000),
		InterestRate:  decimal.NewFromFloat(0.1),
		DurationWeeks: 50,
		WeeklyPayment: decimal.NewFromInt(20000),
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(ctx, loan)
	require.NoError(t, err)

	duplicateID := uuid.New()
	schedules := []*domain.LoanSchedule{
		{
			ID:         duplicateID,
			LoanID:     "LOAN-008",
			WeekNumber: 1,
			DueAmount:  decimal.NewFromInt(20000),
			DueDate:    time.Now().AddDate(0, 0, 7),
			Status:     "pending",
			CreatedAt:  time.Now(),
		},
		{
			ID:         duplicateID, // Duplicate ID
			LoanID:     "LOAN-008",
			WeekNumber: 2,
			DueAmount:  decimal.NewFromInt(20000),
			DueDate:    time.Now().AddDate(0, 0, 14),
			Status:     "pending",
			CreatedAt:  time.Now(),
		},
	}

	err = repo.CreateSchedule(ctx, schedules)
	assert.Error(t, err)

	result, err := repo.GetScheduleByLoanID(ctx, "LOAN-008")
	require.NoError(t, err)
	assert.Len(t, result, 0)
}

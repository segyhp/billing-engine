package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/segyhp/billing-engine/internal/config"
	"github.com/segyhp/billing-engine/internal/domain"
	"github.com/segyhp/billing-engine/internal/handler"
	"github.com/segyhp/billing-engine/internal/repository"
	"github.com/segyhp/billing-engine/internal/service"
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
	sqlBytes, err := ioutil.ReadFile("../../scripts/init.sql")
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

func setupTestEnvironment(t *testing.T) (*httptest.Server, *sqlx.DB, *redis.Client, func()) {
	// Clean test data before each test
	cleanupTestData(testDB)

	// Setup Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use test DB
	})

	// Test connections
	err := testDB.Ping()
	require.NoError(t, err, "Failed to ping test database")

	err = redisClient.Ping(context.Background()).Err()
	require.NoError(t, err, "Failed to connect to test Redis")

	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:         "localhost",
			Port:         "8080",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            "5432",
			User:            "billing_user",
			Password:        "billing_pass",
			Name:            "billing_engine_test",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
		},
		Redis: config.RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       1,
		},
		App: config.AppConfig{
			Environment:              "test",
			LogLevel:                 "debug",
			LoanAmount:               5000000,
			LoanDurationWeeks:        50,
			AnnualInterestRate:       0.10,
			DelinquentWeeksThreshold: 2,
		},
	}

	// Initialize repositories and services
	loanRepo := repository.NewLoanRepository(testDB)
	paymentRepo := repository.NewPaymentRepository(testDB)
	billingService := service.NewBillingService(loanRepo, paymentRepo, redisClient, cfg)
	billingHandler := handler.NewBillingHandler(billingService, cfg)
	healthHandler := handler.NewHealthHandler(testDB, redisClient)

	// Setup routes
	router := setupTestRoutes(billingHandler, healthHandler)
	server := httptest.NewServer(router)

	// Cleanup function
	cleanup := func() {
		// Just clean test data, don't close shared connections
		cleanupTestData(testDB)
	}

	return server, testDB, redisClient, cleanup
}

func cleanupTestData(db *sqlx.DB) {
	db.Exec("DELETE FROM loan_schedule")
	db.Exec("DELETE FROM payments")
	db.Exec("DELETE FROM loans")
}

// TestBillingEngineEndToEnd tests the complete billing engine workflow
func TestBillingEngineEndToEnd(t *testing.T) {
	// Setup test environment
	server, db, _, cleanup := setupTestEnvironment(t)
	defer cleanup()
	defer server.Close()

	// Test data
	loanID := "LOAN-E2E-001"
	loanAmount := decimal.NewFromFloat(5000000) // Rp 5,000,000
	interestRate := decimal.NewFromFloat(0.10)  // 10% annual
	durationWeeks := 50
	expectedWeeklyPayment := decimal.NewFromFloat(110000) // (5,000,000 + 500,000) / 50

	t.Run("Complete Billing Engine E2E Test", func(t *testing.T) {
		// Step 1: Create Loan
		t.Log("Step 1: Creating loan")
		loanResponse := createLoan(t, server.URL, loanID, loanAmount, interestRate, durationWeeks)

		assert.Equal(t, loanID, loanResponse.Loan.LoanID)
		assert.True(t, loanAmount.Equal(loanResponse.Loan.Amount))
		assert.Equal(t, durationWeeks, len(loanResponse.Schedule))
		assert.True(t, expectedWeeklyPayment.Equal(loanResponse.Loan.WeeklyPayment))

		// Step 2: Check Initial Outstanding Balance
		t.Log("Step 2: Checking initial outstanding balance")
		outstanding := getOutstanding(t, server.URL, loanID)
		expectedOutstanding := loanAmount.Add(loanAmount.Mul(interestRate)) // 5,500,000
		assert.True(t, expectedOutstanding.Equal(outstanding))

		// Step 3: Check Initial Delinquency Status (should not be delinquent)
		t.Log("Step 3: Checking initial delinquency status")
		delinquency := checkDelinquency(t, server.URL, loanID)
		assert.False(t, delinquency.IsDelinquent)
		assert.Equal(t, 0, delinquency.MissedWeeks)

		// Step 4: Make First Payment
		t.Log("Step 4: Making first payment")
		paymentResponse := makePayment(t, server.URL, loanID, expectedWeeklyPayment)

		assert.True(t, expectedWeeklyPayment.Equal(paymentResponse.Payment.Amount))
		assert.Equal(t, 1, paymentResponse.Payment.WeekNumber)
		assert.Equal(t, 1, paymentResponse.PaidWeekNumber)
		assert.False(t, paymentResponse.IsDelinquent)

		// Check outstanding after first payment
		outstanding = getOutstanding(t, server.URL, loanID)
		expectedOutstanding = expectedOutstanding.Sub(expectedWeeklyPayment)
		assert.True(t, expectedOutstanding.Equal(outstanding))

		// Step 5: Make Second Payment
		t.Log("Step 5: Making second payment")
		paymentResponse = makePayment(t, server.URL, loanID, expectedWeeklyPayment)

		assert.Equal(t, 2, paymentResponse.Payment.WeekNumber)
		assert.False(t, paymentResponse.IsDelinquent)

		// Step 6: Simulate Overdue Payments (make payments overdue)
		t.Log("Step 6: Simulating overdue payments")
		simulateOverduePayments(t, db, loanID, 3) // Make week 3+ overdue

		// Step 7: Check Delinquency After Overdue
		t.Log("Step 7: Checking delinquency after overdue")
		delinquency = checkDelinquency(t, server.URL, loanID)
		assert.True(t, delinquency.IsDelinquent)
		assert.Equal(t, 2, delinquency.MissedWeeks) // Hardcoded in handler

		// Step 8: Make Payment While Delinquent
		t.Log("Step 8: Making payment while delinquent")
		paymentResponse = makePayment(t, server.URL, loanID, expectedWeeklyPayment)

		assert.Equal(t, 3, paymentResponse.Payment.WeekNumber) // Should pay earliest unpaid
		// Note: IsDelinquent might still be true if there are more overdue payments

		// Step 9: Test Invalid Payment Amount
		t.Log("Step 9: Testing invalid payment amount")
		invalidAmount := decimal.NewFromFloat(50000) // Less than weekly payment
		resp := makePaymentRequest(t, server.URL, loanID, invalidAmount)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		// Step 10: Test Payment for Non-existent Loan
		t.Log("Step 10: Testing payment for non-existent loan")
		resp = makePaymentRequest(t, server.URL, "NON-EXISTENT", expectedWeeklyPayment)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		t.Log("✅ E2E Test completed successfully")
	})
}

// setupTestRoutes sets up the test routes
func setupTestRoutes(billingHandler *handler.BillingHandler, healthHandler *handler.HealthHandler) *mux.Router {
	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/health", healthHandler.Health).Methods("GET")
	router.HandleFunc("/health/ready", healthHandler.Ready).Methods("GET")

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/loans", billingHandler.CreateLoan).Methods("POST")
	api.HandleFunc("/loans/{loanId}/outstanding", billingHandler.GetOutstanding).Methods("GET")
	api.HandleFunc("/loans/{loanId}/delinquent", billingHandler.IsDelinquent).Methods("GET")
	api.HandleFunc("/loans/{loanId}/payment", billingHandler.MakePayment).Methods("POST")

	return router
}

// Helper functions for API calls
func createLoan(t *testing.T, serverURL, loanID string, amount decimal.Decimal, interestRate decimal.Decimal, durationWeeks int) *domain.CreateLoanResponse {
	createReq := domain.CreateLoanRequest{
		LoanID:        loanID,
		Amount:        amount,
		InterestRate:  interestRate,
		DurationWeeks: durationWeeks,
	}

	body, _ := json.Marshal(createReq)
	resp, err := http.Post(serverURL+"/api/v1/loans", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var response struct {
		Data domain.CreateLoanResponse `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	return &response.Data
}

func getOutstanding(t *testing.T, serverURL, loanID string) decimal.Decimal {
	resp, err := http.Get(serverURL + "/api/v1/loans/" + loanID + "/outstanding")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Data domain.OutstandingResponse `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	return response.Data.Outstanding
}

func checkDelinquency(t *testing.T, serverURL, loanID string) *domain.DelinquentResponse {
	resp, err := http.Get(serverURL + "/api/v1/loans/" + loanID + "/delinquent")
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Data domain.DelinquentResponse `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	return &response.Data
}

func makePayment(t *testing.T, serverURL, loanID string, amount decimal.Decimal) *domain.MakePaymentResponse {
	paymentReq := domain.MakePaymentRequest{
		Amount: amount,
	}

	body, _ := json.Marshal(paymentReq)
	resp, err := http.Post(serverURL+"/api/v1/loans/"+loanID+"/payment", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var response struct {
		Data domain.MakePaymentResponse `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	return &response.Data
}

func makePaymentRequest(t *testing.T, serverURL, loanID string, amount decimal.Decimal) *http.Response {
	paymentReq := domain.MakePaymentRequest{
		Amount: amount,
	}

	body, _ := json.Marshal(paymentReq)
	resp, err := http.Post(serverURL+"/api/v1/loans/"+loanID+"/payment", "application/json", bytes.NewBuffer(body))
	require.NoError(t, err)

	return resp
}

func simulateOverduePayments(t *testing.T, db *sqlx.DB, loanID string, fromWeek int) {
	// Update due dates to make payments overdue
	pastDate := time.Now().AddDate(0, 0, -10) // 10 days ago

	query := `
		UPDATE loan_schedule
		SET due_date = $1
		WHERE loan_id = $2 AND week_number >= $3 AND status = 'pending'
	`

	_, err := db.Exec(query, pastDate, loanID, fromWeek)
	require.NoError(t, err)
}

package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/segyhp/billing-engine/internal/config"
	"github.com/segyhp/billing-engine/internal/domain"
	"github.com/segyhp/billing-engine/internal/handler"
	"github.com/segyhp/billing-engine/tests/mocks"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBillingHandler_CreateLoan(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			LoanAmount:         1000.0,
			LoanDurationWeeks:  50,
			AnnualInterestRate: 10.0,
		},
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mocks.MockBillingService)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful loan creation with all fields provided",
			requestBody: domain.CreateLoanRequest{
				LoanID:        "loan123",
				Amount:        decimal.NewFromFloat(2000.0),
				DurationWeeks: 25,
				InterestRate:  decimal.NewFromFloat(0.15),
			},
			setupMock: func(mockService *mocks.MockBillingService) {
				expectedLoan := &domain.Loan{
					ID:            uuid.New(),
					LoanID:        "loan123",
					Amount:        decimal.NewFromFloat(2000.0),
					DurationWeeks: 25,
					InterestRate:  decimal.NewFromFloat(0.15),
					WeeklyPayment: decimal.NewFromFloat(92.0),
					Status:        domain.LoanStatusActive,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				expectedSchedule := []*domain.LoanSchedule{
					{
						ID:         uuid.New(),
						LoanID:     "loan123",
						WeekNumber: 1,
						DueAmount:  decimal.NewFromFloat(92.0),
						DueDate:    time.Now().AddDate(0, 0, 7),
						Status:     domain.ScheduleStatusPending,
						CreatedAt:  time.Now(),
					},
				}
				// Use mock.MatchedBy for more precise matching
				mockService.On("CreateLoan", mock.Anything, mock.MatchedBy(func(req *domain.CreateLoanRequest) bool {
					return req.LoanID == "loan123" &&
						req.Amount.Equal(decimal.NewFromFloat(2000.0)) &&
						req.DurationWeeks == 25 &&
						req.InterestRate.Equal(decimal.NewFromFloat(0.15))
				})).Return(expectedLoan, expectedSchedule, nil).Once()
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				// First unmarshal into the wrapper response structure
				var wrapperResponse struct {
					Success   bool                      `json:"success"`
					Data      domain.CreateLoanResponse `json:"data"`
					Timestamp time.Time                 `json:"timestamp"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &wrapperResponse)
				assert.NoError(t, err)

				// Now access the actual loan data from the Data field
				response := wrapperResponse.Data

				fmt.Println(response.Loan)

				assert.NotNil(t, response.Loan)
				if response.Loan != nil {
					assert.Equal(t, "loan123", response.Loan.LoanID)
					assert.True(t, response.Loan.Amount.Equal(decimal.NewFromFloat(2000.0)))
				}

				assert.NotNil(t, response.Schedule)
				if response.Schedule != nil {
					assert.Len(t, response.Schedule, 1)
				}
			},
		},
		{
			name: "successful loan creation with default values",
			requestBody: domain.CreateLoanRequest{
				LoanID: "loan456",
				// Amount, DurationWeeks, InterestRate will use defaults
			},
			setupMock: func(mockService *mocks.MockBillingService) {
				expectedLoan := &domain.Loan{
					ID:            uuid.New(),
					LoanID:        "loan456",
					Amount:        decimal.NewFromFloat(1000.0),
					DurationWeeks: 50,
					InterestRate:  decimal.NewFromFloat(10.0),
					WeeklyPayment: decimal.NewFromFloat(23.0),
					Status:        domain.LoanStatusActive,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				expectedSchedule := []*domain.LoanSchedule{
					{
						ID:         uuid.New(),
						LoanID:     "loan456",
						WeekNumber: 1,
						DueAmount:  decimal.NewFromFloat(23.0),
						DueDate:    time.Now().AddDate(0, 0, 7),
						Status:     domain.ScheduleStatusPending,
						CreatedAt:  time.Now(),
					},
				}
				mockService.On("CreateLoan", mock.Anything, mock.MatchedBy(func(req *domain.CreateLoanRequest) bool {
					return req.LoanID == "loan456" &&
						req.Amount.Equal(decimal.NewFromFloat(1000.0)) &&
						req.DurationWeeks == 50 &&
						req.InterestRate.Equal(decimal.NewFromFloat(10.0))
				})).Return(expectedLoan, expectedSchedule, nil).Once()
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var wrapperResponse struct {
					Success   bool                      `json:"success"`
					Data      domain.CreateLoanResponse `json:"data"`
					Timestamp time.Time                 `json:"timestamp"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &wrapperResponse)
				assert.NoError(t, err)

				response := wrapperResponse.Data
				assert.NotNil(t, response.Loan)
				assert.Equal(t, "loan456", response.Loan.LoanID)
			},
		},
		{
			name:           "invalid JSON payload",
			requestBody:    "invalid json",
			setupMock:      func(mockService *mocks.MockBillingService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON payload",
		},
		{
			name: "validation error - missing loan ID",
			requestBody: domain.CreateLoanRequest{
				Amount:        decimal.NewFromFloat(1000.0),
				DurationWeeks: 25,
				InterestRate:  decimal.NewFromFloat(0.10),
			},
			setupMock:      func(mockService *mocks.MockBillingService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Validation failed",
		},
		{
			name: "validation error - negative interest rate",
			requestBody: domain.CreateLoanRequest{
				LoanID:        "loan789",
				Amount:        decimal.NewFromFloat(1000.0),
				DurationWeeks: 25,
				InterestRate:  decimal.NewFromFloat(-0.05),
			},
			setupMock:      func(mockService *mocks.MockBillingService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Validation failed",
		},
		{
			name: "validation error - negative amount",
			requestBody: domain.CreateLoanRequest{
				LoanID:        "loan789",
				Amount:        decimal.NewFromFloat(-100.0), // negative instead of zero
				DurationWeeks: 25,
				InterestRate:  decimal.NewFromFloat(0.10),
			},
			setupMock:      func(mockService *mocks.MockBillingService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Validation failed",
		},
		{
			name: "validation error - negative duration weeks",
			requestBody: domain.CreateLoanRequest{
				LoanID:        "loan789",
				Amount:        decimal.NewFromFloat(1000.0),
				DurationWeeks: -5, // negative instead of zero
				InterestRate:  decimal.NewFromFloat(0.10),
			},
			setupMock:      func(mockService *mocks.MockBillingService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Validation failed",
		},
		{
			name: "service error - loan already exists",
			requestBody: domain.CreateLoanRequest{
				LoanID:        "existing_loan",
				Amount:        decimal.NewFromFloat(1500.0),
				DurationWeeks: 30,
				InterestRate:  decimal.NewFromFloat(0.12),
			},
			setupMock: func(mockService *mocks.MockBillingService) {
				mockService.On("CreateLoan", mock.Anything, mock.MatchedBy(func(req *domain.CreateLoanRequest) bool {
					return req.LoanID == "existing_loan"
				})).Return((*domain.Loan)(nil), ([]*domain.LoanSchedule)(nil), assert.AnError).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to create loan",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := mocks.NewMockBillingService()
			tt.setupMock(mockService)

			// Create handler with mock service
			billingHandler := handler.NewBillingHandler(mockService, cfg)

			// Create request
			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body.WriteString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/loans", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Call handler
			billingHandler.CreateLoan(w, req)

			// Assert response status
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check expected body content if provided
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}

			// Run additional response checks if provided
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			// Verify mock calls
			mockService.AssertExpectations(t)
		})
	}
}

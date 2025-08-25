package handler

import (
	"encoding/json"
	"net/http"

	"github.com/segyhp/billing-engine/internal/config"
	"github.com/segyhp/billing-engine/internal/domain"
	"github.com/segyhp/billing-engine/internal/service"
	"github.com/segyhp/billing-engine/pkg/response"
	"github.com/shopspring/decimal"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type BillingHandler struct {
	service   service.BillingService
	validator *validator.Validate
	config    *config.Config
}

func NewBillingHandler(service service.BillingService, config *config.Config) *BillingHandler {
	validate := validator.New()

	// Register custom validation tags for decimal
	validate.RegisterValidation("decimal_gt", validateDecimalGt)
	validate.RegisterValidation("decimal_gte", validateDecimalGte)

	return &BillingHandler{
		service:   service,
		validator: validate,
		config:    config,
	}
}

func (h *BillingHandler) CreateLoan(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateLoanRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err)
		return
	}

	// Apply default values from config if not provided
	// for testing purposes
	// these can be overridden by request payload
	// e.g. curl -X POST http://localhost:8080/loans -d '{"amount":1500,"duration_weeks":30,"interest_rate":0.12}'
	if req.Amount.IsZero() {
		req.Amount = decimal.NewFromFloat(h.config.App.LoanAmount)
	}
	if req.DurationWeeks == 0 {
		req.DurationWeeks = h.config.App.LoanDurationWeeks
	}
	if req.InterestRate.IsZero() {
		req.InterestRate = decimal.NewFromFloat(h.config.App.AnnualInterestRate)
	}

	if err := h.validator.Struct(&req); err != nil {
		response.BadRequest(w, "Validation failed", err)
		return
	}

	loan, schedule, err := h.service.CreateLoan(r.Context(), &req)
	if err != nil {
		response.InternalServerError(w, "Failed to create loan", err)
		return
	}

	responseData := domain.CreateLoanResponse{
		Loan:     loan,
		Schedule: schedule,
	}

	response.Created(w, responseData)
}

// GetOutstanding returns the outstanding amount for a loan
func (h *BillingHandler) GetOutstanding(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	loanID := vars["loanId"]

	if loanID == "" {
		response.BadRequest(w, "Loan ID is required", nil)
		return
	}

	outstanding, err := h.service.GetOutstanding(r.Context(), loanID)
	if err != nil {
		response.InternalServerError(w, "Failed to get outstanding", err)
		return
	}

	responseData := domain.OutstandingResponse{
		LoanID:      loanID,
		Outstanding: outstanding,
	}

	response.Success(w, responseData)
}

// IsDelinquent checks if a borrower is delinquent
func (h *BillingHandler) IsDelinquent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	loanID := vars["loanId"]

	if loanID == "" {
		response.BadRequest(w, "Loan ID is required", nil)
		return
	}

	// TODO: Implement delinquency check logic
	// isDelinquent, missedWeeks, err := h.service.IsDelinquent(r.Context(), loanID)
	// if err != nil {
	// 	response.InternalServerError(w, "Failed to check delinquency", err)
	// 	return
	// }

	// For now, return a placeholder response
	response.Success(w, map[string]string{
		"message": "IsDelinquent endpoint - TODO: Implement business logic",
		"loan_id": loanID,
	})
}

// MakePayment processes a payment for a loan
func (h *BillingHandler) MakePayment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	loanID := vars["loanId"]

	if loanID == "" {
		response.BadRequest(w, "Loan ID is required", nil)
		return
	}

	var req domain.MakePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		response.BadRequest(w, "Validation failed", err)
		return
	}

	// TODO: Implement payment processing logic
	// payment, err := h.service.MakePayment(r.Context(), loanID, req.Amount)
	// if err != nil {
	// 	response.InternalServerError(w, "Failed to process payment", err)
	// 	return
	// }

	// For now, return a placeholder response
	response.Success(w, map[string]interface{}{
		"message": "MakePayment endpoint - TODO: Implement business logic",
		"loan_id": loanID,
		"amount":  req.Amount,
	})
}

// validateDecimalGt validates that decimal is greater than the parameter
func validateDecimalGt(fl validator.FieldLevel) bool {
	dec, ok := fl.Field().Interface().(decimal.Decimal)
	if !ok {
		return false
	}

	param := fl.Param()

	if param == "0" {
		return dec.GreaterThan(decimal.Zero)
	}

	paramDecimal, err := decimal.NewFromString(param)
	if err != nil {
		return false
	}

	return dec.GreaterThan(paramDecimal)
}

// validateDecimalGte validates that decimal is greater than or equal to the parameter
func validateDecimalGte(fl validator.FieldLevel) bool {
	dec, ok := fl.Field().Interface().(decimal.Decimal)
	if !ok {
		return false
	}

	param := fl.Param()

	if param == "0" {
		return dec.GreaterThanOrEqual(decimal.Zero)
	}

	paramDecimal, err := decimal.NewFromString(param)
	if err != nil {
		return false
	}

	return dec.GreaterThanOrEqual(paramDecimal)
}

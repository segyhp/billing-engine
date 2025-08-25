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

	isDelinquent, err := h.service.IsDelinquent(r.Context(), loanID)
	if err != nil {
		response.InternalServerError(w, "Failed to check delinquency", err)
		return
	}

	// Calculate missed weeks by checking the service logic
	// For now, we'll set it based on delinquency status
	missedWeeks := 0
	if isDelinquent {
		missedWeeks = 2 // Minimum threshold for delinquency
	}

	responseData := domain.DelinquentResponse{
		LoanID:       loanID,
		IsDelinquent: isDelinquent,
		MissedWeeks:  missedWeeks,
	}

	response.Success(w, responseData)
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

	// Set the loan ID from URL params
	req.LoanID = loanID

	if err := h.validator.Struct(&req); err != nil {
		response.BadRequest(w, "Validation failed", err)
		return
	}

	payment, err := h.service.MakePayment(r.Context(), req)
	if err != nil {
		response.InternalServerError(w, "Failed to process payment", err)
		return
	}

	// Get updated outstanding balance after payment
	outstanding, err := h.service.GetOutstanding(r.Context(), loanID)
	if err != nil {
		response.InternalServerError(w, "Failed to get outstanding balance", err)
		return
	}

	// Check if borrower is still delinquent after payment
	isDelinquent, err := h.service.IsDelinquent(r.Context(), loanID)
	if err != nil {
		response.InternalServerError(w, "Failed to check delinquency status", err)
		return
	}

	responseData := domain.MakePaymentResponse{
		Payment:        payment,
		Outstanding:    outstanding,
		IsDelinquent:   isDelinquent,
		PaidWeekNumber: payment.WeekNumber,
	}

	response.Success(w, responseData)
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

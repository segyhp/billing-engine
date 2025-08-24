package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/segyhp/billing-engine/internal/service"
)

type BillingHandler struct {
	service   *service.BillingService
	validator *validator.Validate
}

func NewBillingHandler(service *service.BillingService) *BillingHandler {
	return &BillingHandler{
		service:   service,
		validator: validator.New(),
	}
}

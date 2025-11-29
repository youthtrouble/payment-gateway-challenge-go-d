package models

import (
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain"
)

type PostPaymentRequest struct {
	CardNumber  string `json:"card_number" example:"2222405343248877" validate:"required,min=14,max=19,numeric"`
	ExpiryMonth int    `json:"expiry_month" example:"12" validate:"required,min=1,max=12"`
	ExpiryYear  int    `json:"expiry_year" example:"2026" validate:"required"`
	Currency    string `json:"currency" example:"GBP" validate:"required,len=3,oneof=USD GBP EUR"`
	Amount      int    `json:"amount" example:"100" validate:"required,min=1"`
	CVV         string `json:"cvv" example:"123" validate:"required,min=3,max=4,numeric"`
}

type PostPaymentResponse struct {
	ID                 string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status             string `json:"status" example:"Authorized" enums:"Authorized,Declined,Rejected"`
	CardNumberLastFour string `json:"card_number_last_four" example:"8877"`
	ExpiryMonth        int    `json:"expiry_month" example:"12"`
	ExpiryYear         int    `json:"expiry_year" example:"2026"`
	Currency           string `json:"currency" example:"GBP"`
	Amount             int    `json:"amount" example:"100"`
}

type GetPaymentResponse struct {
	ID                 string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status             string `json:"status" example:"Authorized" enums:"Authorized,Declined"`
	CardNumberLastFour string `json:"card_number_last_four" example:"8877"`
	ExpiryMonth        int    `json:"expiry_month" example:"12"`
	ExpiryYear         int    `json:"expiry_year" example:"2026"`
	Currency           string `json:"currency" example:"GBP"`
	Amount             int    `json:"amount" example:"100"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"card number must be between 14-19 digits"` // Error message
}

func (r *PostPaymentRequest) ToDomainPayment() (*domain.Payment, error) {
	card := domain.Card{
		Number:      r.CardNumber,
		ExpiryMonth: r.ExpiryMonth,
		ExpiryYear:  r.ExpiryYear,
		CVV:         r.CVV,
	}

	return domain.NewPayment(card, r.Currency, r.Amount)
}

func FromDomainPayment(payment *domain.Payment) *PostPaymentResponse {

	lastFour := payment.Card.GetLastFourDigits()

	return &PostPaymentResponse{
		ID:                 payment.ID,
		Status:             string(payment.Status),
		CardNumberLastFour: lastFour,
		ExpiryMonth:        payment.Card.ExpiryMonth,
		ExpiryYear:         payment.Card.ExpiryYear,
		Currency:           payment.Currency,
		Amount:             payment.Amount,
	}
}

func ToGetPaymentResponse(payment *domain.Payment) *GetPaymentResponse {
	lastFour := payment.Card.GetLastFourDigits()

	return &GetPaymentResponse{
		ID:                 payment.ID,
		Status:             string(payment.Status),
		CardNumberLastFour: lastFour,
		ExpiryMonth:        payment.Card.ExpiryMonth,
		ExpiryYear:         payment.Card.ExpiryYear,
		Currency:           payment.Currency,
		Amount:             payment.Amount,
	}
}

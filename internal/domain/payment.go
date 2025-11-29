package domain

import (
	"strings"
)

type PaymentStatus string

const (
	// StatusAuthorized means the payment was authorized by the bank
	StatusAuthorized PaymentStatus = "Authorized"
	// StatusDeclined means the payment was declined by the bank
	StatusDeclined PaymentStatus = "Declined"
	// StatusRejected means the payment was rejected due to validation errors
	StatusRejected PaymentStatus = "Rejected"
)

var supportedCurrencies = map[string]bool{
	"USD": true,
	"GBP": true,
	"EUR": true,
}

type Payment struct {
	ID       string
	Card     Card
	Currency string
	Amount   int
	Status   PaymentStatus
}

func NewPayment(card Card, currency string, amount int) (*Payment, error) {
	p := &Payment{
		Card:     card,
		Currency: currency,
		Amount:   amount,
		Status:   StatusRejected, // Default to rejected until validated
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Payment) Validate() error {

	if err := p.Card.Validate(); err != nil {
		return err
	}

	if err := p.validateCurrency(); err != nil {
		return err
	}

	if err := p.validateAmount(); err != nil {
		return err
	}

	return nil
}

// validateCurrency ensures currency meets requirements:
func (p *Payment) validateCurrency() error {
	if p.Currency == "" {
		return ErrCurrencyRequired
	}

	currency := strings.ToUpper(p.Currency)

	if len(currency) != 3 {
		return ErrCurrencyInvalid
	}

	if !supportedCurrencies[currency] {
		return ErrCurrencyInvalid
	}

	p.Currency = currency

	return nil
}

// validateAmount ensures amount is valid:
func (p *Payment) validateAmount() error {
	if p.Amount <= 0 {
		return ErrAmountInvalid
	}

	return nil
}

func (p *Payment) SetAuthorized() {
	p.Status = StatusAuthorized
}

func (p *Payment) SetDeclined() {
	p.Status = StatusDeclined
}

func (p *Payment) SetRejected() {
	p.Status = StatusRejected
}

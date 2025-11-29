package domain

import "errors"

// Domain-specific errors for validation and business logic
var (
	// Card validation errors
	ErrCardNumberRequired    = errors.New("card number is required")
	ErrCardNumberInvalid     = errors.New("card number must be between 14-19 digits")
	ErrCardNumberNotNumeric  = errors.New("card number must only contain numeric characters")
	ErrCVVRequired           = errors.New("CVV is required")
	ErrCVVInvalid            = errors.New("CVV must be 3-4 digits")
	ErrCVVNotNumeric         = errors.New("CVV must only contain numeric characters")
	ErrExpiryMonthRequired   = errors.New("expiry month is required")
	ErrExpiryMonthInvalid    = errors.New("expiry month must be between 1-12")
	ErrExpiryYearRequired    = errors.New("expiry year is required")
	ErrExpiryDateInPast      = errors.New("expiry date must be in the future")

	// Payment validation errors
	ErrCurrencyRequired = errors.New("currency is required")
	ErrCurrencyInvalid  = errors.New("currency must be a valid 3-character ISO code (USD, GBP, EUR)")
	ErrAmountRequired   = errors.New("amount is required")
	ErrAmountInvalid    = errors.New("amount must be a positive integer")

	// Business logic errors
	ErrPaymentNotFound = errors.New("payment not found")
)

package domain

import (
	"strconv"
	"time"
)

type Card struct {
	Number      string
	ExpiryMonth int
	ExpiryYear  int
	CVV         string
}

func (c *Card) Validate() error {
	if err := c.validateCardNumber(); err != nil {
		return err
	}

	if err := c.validateExpiry(); err != nil {
		return err
	}

	if err := c.validateCVV(); err != nil {
		return err
	}

	return nil
}

// validateCardNumber ensures card number meets requirements:
func (c *Card) validateCardNumber() error {
	if c.Number == "" {
		return ErrCardNumberRequired
	}

	length := len(c.Number)
	if length < 14 || length > 19 {
		return ErrCardNumberInvalid
	}

	if !isNumeric(c.Number) {
		return ErrCardNumberNotNumeric
	}

	return nil
}

// validateExpiry ensures expiry date is valid and in the future
func (c *Card) validateExpiry() error {

	if c.ExpiryMonth == 0 {
		return ErrExpiryMonthRequired
	}

	if c.ExpiryMonth < 1 || c.ExpiryMonth > 12 {
		return ErrExpiryMonthInvalid
	}

	if c.ExpiryYear == 0 {
		return ErrExpiryYearRequired
	}

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	// If expiry year is less than current year, it's expired
	if c.ExpiryYear < currentYear {
		return ErrExpiryDateInPast
	}

	// If expiry year equals current year, check the month
	if c.ExpiryYear == currentYear && c.ExpiryMonth < currentMonth {
		return ErrExpiryDateInPast
	}

	return nil
}

// validateCVV ensures CVV meets requirements:
func (c *Card) validateCVV() error {
	if c.CVV == "" {
		return ErrCVVRequired
	}

	length := len(c.CVV)
	if length < 3 || length > 4 {
		return ErrCVVInvalid
	}

	if !isNumeric(c.CVV) {
		return ErrCVVNotNumeric
	}

	return nil
}

func (c *Card) GetLastFourDigits() string {
	if len(c.Number) < 4 {
		return c.Number
	}
	return c.Number[len(c.Number)-4:]
}

func isNumeric(s string) bool {
	_, err := strconv.ParseUint(s, 10, 64)
	return err == nil
}

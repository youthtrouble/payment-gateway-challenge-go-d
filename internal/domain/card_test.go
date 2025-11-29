package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCard_ValidateCardNumber(t *testing.T) {
	tests := []struct {
		name        string
		cardNumber  string
		expectError error
	}{
		{
			name:        "valid 16-digit card",
			cardNumber:  "1234567890123456",
			expectError: nil,
		},
		{
			name:        "valid 14-digit card (minimum)",
			cardNumber:  "12345678901234",
			expectError: nil,
		},
		{
			name:        "valid 19-digit card (maximum)",
			cardNumber:  "1234567890123456789",
			expectError: nil,
		},
		{
			name:        "empty card number",
			cardNumber:  "",
			expectError: ErrCardNumberRequired,
		},
		{
			name:        "card number too short (13 digits)",
			cardNumber:  "1234567890123",
			expectError: ErrCardNumberInvalid,
		},
		{
			name:        "card number too long (20 digits)",
			cardNumber:  "12345678901234567890",
			expectError: ErrCardNumberInvalid,
		},
		{
			name:        "card number with letters",
			cardNumber:  "123456789012345A",
			expectError: ErrCardNumberNotNumeric,
		},
		{
			name:        "card number with spaces",
			cardNumber:  "1234 5678 9012 3456",
			expectError: ErrCardNumberNotNumeric,
		},
		{
			name:        "card number with special characters",
			cardNumber:  "1234-5678-9012-3456",
			expectError: ErrCardNumberNotNumeric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := Card{
				Number:      tt.cardNumber,
				ExpiryMonth: 12,
				ExpiryYear:  time.Now().Year() + 1,
				CVV:         "123",
			}

			err := card.validateCardNumber()
			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCard_ValidateExpiry(t *testing.T) {
	currentYear := time.Now().Year()
	currentMonth := int(time.Now().Month())

	tests := []struct {
		name        string
		expiryMonth int
		expiryYear  int
		expectError error
	}{
		{
			name:        "valid future date",
			expiryMonth: 12,
			expiryYear:  currentYear + 1,
			expectError: nil,
		},
		{
			name:        "valid current year, future month",
			expiryMonth: 12,
			expiryYear:  currentYear,
			expectError: nil,
		},
		{
			name:        "month not provided",
			expiryMonth: 0,
			expiryYear:  currentYear + 1,
			expectError: ErrExpiryMonthRequired,
		},
		{
			name:        "year not provided",
			expiryMonth: 12,
			expiryYear:  0,
			expectError: ErrExpiryYearRequired,
		},
		{
			name:        "month less than 1",
			expiryMonth: 0,
			expiryYear:  currentYear + 1,
			expectError: ErrExpiryMonthRequired,
		},
		{
			name:        "month greater than 12",
			expiryMonth: 13,
			expiryYear:  currentYear + 1,
			expectError: ErrExpiryMonthInvalid,
		},
		{
			name:        "past year",
			expiryMonth: 12,
			expiryYear:  currentYear - 1,
			expectError: ErrExpiryDateInPast,
		},
		{
			name:        "current year, past month",
			expiryMonth: currentMonth - 1,
			expiryYear:  currentYear,
			expectError: func() error {
				if currentMonth == 1 {
					// If current month is January, last month was previous year
					return nil
				}
				return ErrExpiryDateInPast
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := Card{
				Number:      "1234567890123456",
				ExpiryMonth: tt.expiryMonth,
				ExpiryYear:  tt.expiryYear,
				CVV:         "123",
			}

			err := card.validateExpiry()
			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCard_ValidateCVV(t *testing.T) {
	tests := []struct {
		name        string
		cvv         string
		expectError error
	}{
		{
			name:        "valid 3-digit CVV",
			cvv:         "123",
			expectError: nil,
		},
		{
			name:        "valid 4-digit CVV",
			cvv:         "1234",
			expectError: nil,
		},
		{
			name:        "empty CVV",
			cvv:         "",
			expectError: ErrCVVRequired,
		},
		{
			name:        "CVV too short (2 digits)",
			cvv:         "12",
			expectError: ErrCVVInvalid,
		},
		{
			name:        "CVV too long (5 digits)",
			cvv:         "12345",
			expectError: ErrCVVInvalid,
		},
		{
			name:        "CVV with letters",
			cvv:         "12A",
			expectError: ErrCVVNotNumeric,
		},
		{
			name:        "CVV with special characters",
			cvv:         "12*",
			expectError: ErrCVVNotNumeric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := Card{
				Number:      "1234567890123456",
				ExpiryMonth: 12,
				ExpiryYear:  time.Now().Year() + 1,
				CVV:         tt.cvv,
			}

			err := card.validateCVV()
			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCard_Validate(t *testing.T) {
	currentYear := time.Now().Year()

	tests := []struct {
		name        string
		card        Card
		expectError error
	}{
		{
			name: "valid card",
			card: Card{
				Number:      "1234567890123456",
				ExpiryMonth: 12,
				ExpiryYear:  currentYear + 1,
				CVV:         "123",
			},
			expectError: nil,
		},
		{
			name: "invalid card number",
			card: Card{
				Number:      "123", // too short
				ExpiryMonth: 12,
				ExpiryYear:  currentYear + 1,
				CVV:         "123",
			},
			expectError: ErrCardNumberInvalid,
		},
		{
			name: "invalid expiry",
			card: Card{
				Number:      "1234567890123456",
				ExpiryMonth: 13, // invalid month
				ExpiryYear:  currentYear + 1,
				CVV:         "123",
			},
			expectError: ErrExpiryMonthInvalid,
		},
		{
			name: "invalid CVV",
			card: Card{
				Number:      "1234567890123456",
				ExpiryMonth: 12,
				ExpiryYear:  currentYear + 1,
				CVV:         "12", // too short
			},
			expectError: ErrCVVInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()
			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCard_GetLastFourDigits(t *testing.T) {
	tests := []struct {
		name       string
		cardNumber string
		expected   string
	}{
		{
			name:       "16-digit card",
			cardNumber: "1234567890123456",
			expected:   "3456",
		},
		{
			name:       "14-digit card",
			cardNumber: "12345678901234",
			expected:   "1234",
		},
		{
			name:       "card less than 4 digits (edge case)",
			cardNumber: "123",
			expected:   "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := Card{Number: tt.cardNumber}
			result := card.GetLastFourDigits()
			assert.Equal(t, tt.expected, result)
		})
	}
}

package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPayment_ValidateCurrency(t *testing.T) {
	tests := []struct {
		name        string
		currency    string
		expectError error
		expected    string
	}{
		{
			name:        "valid USD",
			currency:    "USD",
			expectError: nil,
			expected:    "USD",
		},
		{
			name:        "valid GBP",
			currency:    "GBP",
			expectError: nil,
			expected:    "GBP",
		},
		{
			name:        "valid EUR",
			currency:    "EUR",
			expectError: nil,
			expected:    "EUR",
		},
		{
			name:        "valid lowercase usd (should normalize)",
			currency:    "usd",
			expectError: nil,
			expected:    "USD",
		},
		{
			name:        "empty currency",
			currency:    "",
			expectError: ErrCurrencyRequired,
		},
		{
			name:        "invalid currency code",
			currency:    "JPY",
			expectError: ErrCurrencyInvalid,
		},
		{
			name:        "currency too short",
			currency:    "US",
			expectError: ErrCurrencyInvalid,
		},
		{
			name:        "currency too long",
			currency:    "USDD",
			expectError: ErrCurrencyInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := &Payment{
				Currency: tt.currency,
				Amount:   100,
				Card: Card{
					Number:      "1234567890123456",
					ExpiryMonth: 12,
					ExpiryYear:  time.Now().Year() + 1,
					CVV:         "123",
				},
			}

			err := payment.validateCurrency()
			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, payment.Currency)
			}
		})
	}
}

func TestPayment_ValidateAmount(t *testing.T) {
	tests := []struct {
		name        string
		amount      int
		expectError error
	}{
		{
			name:        "valid amount (1 cent)",
			amount:      1,
			expectError: nil,
		},
		{
			name:        "valid amount (100 dollars)",
			amount:      10000,
			expectError: nil,
		},
		{
			name:        "zero amount",
			amount:      0,
			expectError: ErrAmountInvalid,
		},
		{
			name:        "negative amount",
			amount:      -100,
			expectError: ErrAmountInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := &Payment{
				Amount:   tt.amount,
				Currency: "USD",
				Card: Card{
					Number:      "1234567890123456",
					ExpiryMonth: 12,
					ExpiryYear:  time.Now().Year() + 1,
					CVV:         "123",
				},
			}

			err := payment.validateAmount()
			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewPayment(t *testing.T) {
	currentYear := time.Now().Year()

	tests := []struct {
		name        string
		card        Card
		currency    string
		amount      int
		expectError error
	}{
		{
			name: "valid payment",
			card: Card{
				Number:      "1234567890123456",
				ExpiryMonth: 12,
				ExpiryYear:  currentYear + 1,
				CVV:         "123",
			},
			currency:    "USD",
			amount:      1000,
			expectError: nil,
		},
		{
			name: "invalid card",
			card: Card{
				Number:      "123", // too short
				ExpiryMonth: 12,
				ExpiryYear:  currentYear + 1,
				CVV:         "123",
			},
			currency:    "USD",
			amount:      1000,
			expectError: ErrCardNumberInvalid,
		},
		{
			name: "invalid currency",
			card: Card{
				Number:      "1234567890123456",
				ExpiryMonth: 12,
				ExpiryYear:  currentYear + 1,
				CVV:         "123",
			},
			currency:    "JPY",
			amount:      1000,
			expectError: ErrCurrencyInvalid,
		},
		{
			name: "invalid amount",
			card: Card{
				Number:      "1234567890123456",
				ExpiryMonth: 12,
				ExpiryYear:  currentYear + 1,
				CVV:         "123",
			},
			currency:    "USD",
			amount:      0,
			expectError: ErrAmountInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment, err := NewPayment(tt.card, tt.currency, tt.amount)
			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
				assert.Nil(t, payment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payment)
				assert.Equal(t, StatusRejected, payment.Status) // Default status
			}
		})
	}
}

func TestPayment_StatusMethods(t *testing.T) {
	payment := &Payment{
		Status: StatusRejected,
	}

	// Test SetAuthorized
	payment.SetAuthorized()
	assert.Equal(t, StatusAuthorized, payment.Status)

	// Test SetDeclined
	payment.SetDeclined()
	assert.Equal(t, StatusDeclined, payment.Status)

	// Test SetRejected
	payment.SetRejected()
	assert.Equal(t, StatusRejected, payment.Status)
}

func TestPayment_Validate(t *testing.T) {
	currentYear := time.Now().Year()

	tests := []struct {
		name        string
		payment     Payment
		expectError error
	}{
		{
			name: "fully valid payment",
			payment: Payment{
				Card: Card{
					Number:      "1234567890123456",
					ExpiryMonth: 12,
					ExpiryYear:  currentYear + 1,
					CVV:         "123",
				},
				Currency: "USD",
				Amount:   1000,
			},
			expectError: nil,
		},
		{
			name: "invalid - bad card",
			payment: Payment{
				Card: Card{
					Number:      "ABC", // non-numeric
					ExpiryMonth: 12,
					ExpiryYear:  currentYear + 1,
					CVV:         "123",
				},
				Currency: "USD",
				Amount:   1000,
			},
			expectError: ErrCardNumberInvalid,
		},
		{
			name: "invalid - bad currency",
			payment: Payment{
				Card: Card{
					Number:      "1234567890123456",
					ExpiryMonth: 12,
					ExpiryYear:  currentYear + 1,
					CVV:         "123",
				},
				Currency: "",
				Amount:   1000,
			},
			expectError: ErrCurrencyRequired,
		},
		{
			name: "invalid - bad amount",
			payment: Payment{
				Card: Card{
					Number:      "1234567890123456",
					ExpiryMonth: 12,
					ExpiryYear:  currentYear + 1,
					CVV:         "123",
				},
				Currency: "USD",
				Amount:   -100,
			},
			expectError: ErrAmountInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payment.Validate()
			if tt.expectError != nil {
				assert.Equal(t, tt.expectError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPBankClient_ProcessPayment_Success_Authorized(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/payments", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req BankRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "1234567890123456", req.CardNumber)
		assert.Equal(t, "12/2025", req.ExpiryDate)
		assert.Equal(t, "USD", req.Currency)
		assert.Equal(t, 1000, req.Amount)
		assert.Equal(t, "123", req.CVV)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(BankResponse{
			Authorized:        true,
			AuthorizationCode: "auth-123",
		})
	}))
	defer server.Close()

	client := NewHTTPBankClient(server.URL)

	payment := &domain.Payment{
		Card: domain.Card{
			Number:      "1234567890123456",
			ExpiryMonth: 12,
			ExpiryYear:  2025,
			CVV:         "123",
		},
		Currency: "USD",
		Amount:   1000,
	}

	resp, err := client.ProcessPayment(payment)

	require.NoError(t, err)
	assert.True(t, resp.Authorized)
	assert.Equal(t, "auth-123", resp.AuthorizationCode)
}

func TestHTTPBankClient_ProcessPayment_Success_Declined(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(BankResponse{
			Authorized:        false,
			AuthorizationCode: "",
		})
	}))
	defer server.Close()

	client := NewHTTPBankClient(server.URL)

	payment := &domain.Payment{
		Card: domain.Card{
			Number:      "1234567890123456",
			ExpiryMonth: 12,
			ExpiryYear:  2025,
			CVV:         "123",
		},
		Currency: "USD",
		Amount:   1000,
	}

	resp, err := client.ProcessPayment(payment)

	require.NoError(t, err)
	assert.False(t, resp.Authorized)
	assert.Empty(t, resp.AuthorizationCode)
}

func TestHTTPBankClient_ProcessPayment_BadRequest(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid card number"))
	}))
	defer server.Close()

	client := NewHTTPBankClient(server.URL)

	payment := &domain.Payment{
		Card: domain.Card{
			Number:      "1234567890123456",
			ExpiryMonth: 12,
			ExpiryYear:  2025,
			CVV:         "123",
		},
		Currency: "USD",
		Amount:   1000,
	}

	resp, err := client.ProcessPayment(payment)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "bank rejected request")
}

func TestHTTPBankClient_ProcessPayment_ServiceUnavailable(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewHTTPBankClient(server.URL)

	payment := &domain.Payment{
		Card: domain.Card{
			Number:      "1234567890123456",
			ExpiryMonth: 12,
			ExpiryYear:  2025,
			CVV:         "123",
		},
		Currency: "USD",
		Amount:   1000,
	}

	resp, err := client.ProcessPayment(payment)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "bank service unavailable")
}

func TestHTTPBankClient_ProcessPayment_InvalidJSON(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewHTTPBankClient(server.URL)

	payment := &domain.Payment{
		Card: domain.Card{
			Number:      "1234567890123456",
			ExpiryMonth: 12,
			ExpiryYear:  2025,
			CVV:         "123",
		},
		Currency: "USD",
		Amount:   1000,
	}

	resp, err := client.ProcessPayment(payment)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to unmarshal bank response")
}

func TestHTTPBankClient_ConvertToBankRequest(t *testing.T) {
	client := NewHTTPBankClient("http://localhost:8081")

	payment := &domain.Payment{
		Card: domain.Card{
			Number:      "1234567890123456",
			ExpiryMonth: 4,
			ExpiryYear:  2025,
			CVV:         "123",
		},
		Currency: "GBP",
		Amount:   500,
	}

	bankReq := client.convertTobankRequest(payment)

	assert.Equal(t, "1234567890123456", bankReq.CardNumber)
	assert.Equal(t, "04/2025", bankReq.ExpiryDate) // Month should be zero-padded
	assert.Equal(t, "GBP", bankReq.Currency)
	assert.Equal(t, 500, bankReq.Amount)
	assert.Equal(t, "123", bankReq.CVV)
}

func TestHTTPBankClient_ExpiryDateFormatting(t *testing.T) {
	client := NewHTTPBankClient("http://localhost:8081")

	tests := []struct {
		name     string
		month    int
		year     int
		expected string
	}{
		{
			name:     "single digit month",
			month:    1,
			year:     2025,
			expected: "01/2025",
		},
		{
			name:     "double digit month",
			month:    12,
			year:     2025,
			expected: "12/2025",
		},
		{
			name:     "April",
			month:    4,
			year:     2026,
			expected: "04/2026",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := &domain.Payment{
				Card: domain.Card{
					Number:      "1234567890123456",
					ExpiryMonth: tt.month,
					ExpiryYear:  tt.year,
					CVV:         "123",
				},
				Currency: "USD",
				Amount:   100,
			}

			bankReq := client.convertTobankRequest(payment)
			assert.Equal(t, tt.expected, bankReq.ExpiryDate)
		})
	}
}

func TestHTTPBankClient_Timeout(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(15 * time.Second) // Longer than client timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPBankClient(server.URL)

	payment := &domain.Payment{
		Card: domain.Card{
			Number:      "1234567890123456",
			ExpiryMonth: 12,
			ExpiryYear:  2025,
			CVV:         "123",
		},
		Currency: "USD",
		Amount:   1000,
	}

	resp, err := client.ProcessPayment(payment)

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to send request to bank")
}

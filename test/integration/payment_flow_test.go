package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/api"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPaymentFlow_Authorized tests the full payment flow with a card ending in odd number (authorized)
func TestPaymentFlow_Authorized(t *testing.T) {
	testAPI := api.New()

	futureYear := time.Now().Year() + 1

	// Step 1: Create a payment with card ending in odd number (should be authorized)
	reqBody := models.PostPaymentRequest{
		CardNumber:  "2222405343248877", // Ends in 7 (odd) - will be authorized
		ExpiryMonth: 4,
		ExpiryYear:  futureYear,
		Currency:    "GBP",
		Amount:      100,
		CVV:         "123",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	testAPI.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var postResp models.PostPaymentResponse
	err := json.NewDecoder(w.Body).Decode(&postResp)
	require.NoError(t, err)

	assert.NotEmpty(t, postResp.ID)
	assert.Equal(t, "Authorized", postResp.Status)
	assert.Equal(t, "8877", postResp.CardNumberLastFour)
	assert.Equal(t, 4, postResp.ExpiryMonth)
	assert.Equal(t, futureYear, postResp.ExpiryYear)
	assert.Equal(t, "GBP", postResp.Currency)
	assert.Equal(t, 100, postResp.Amount)

	// Step 2: Retrieve the payment by ID
	paymentID := postResp.ID
	getReq := httptest.NewRequest(http.MethodGet, "/api/payments/"+paymentID, nil)
	getW := httptest.NewRecorder()

	testAPI.Router().ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusOK, getW.Code)

	var getResp models.GetPaymentResponse
	err = json.NewDecoder(getW.Body).Decode(&getResp)
	require.NoError(t, err)

	assert.Equal(t, paymentID, getResp.ID)
	assert.Equal(t, "Authorized", getResp.Status)
	assert.Equal(t, "8877", getResp.CardNumberLastFour)
	assert.Equal(t, 4, getResp.ExpiryMonth)
	assert.Equal(t, futureYear, getResp.ExpiryYear)
	assert.Equal(t, "GBP", getResp.Currency)
	assert.Equal(t, 100, getResp.Amount)
}

// TestPaymentFlow_Declined tests the full payment flow with a card ending in even number (declined)
func TestPaymentFlow_Declined(t *testing.T) {
	testAPI := api.New()
	futureYear := time.Now().Year() + 1

	reqBody := models.PostPaymentRequest{
		CardNumber:  "2222405343248878", // Ends in 8 (even) - will be declined
		ExpiryMonth: 4,
		ExpiryYear:  futureYear,
		Currency:    "GBP",
		Amount:      100,
		CVV:         "123",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	testAPI.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var postResp models.PostPaymentResponse
	err := json.NewDecoder(w.Body).Decode(&postResp)
	require.NoError(t, err)

	assert.NotEmpty(t, postResp.ID)
	assert.Equal(t, "Declined", postResp.Status) // Should be declined
	assert.Equal(t, "8878", postResp.CardNumberLastFour)

	// Retrieve the declined payment
	getReq := httptest.NewRequest(http.MethodGet, "/api/payments/"+postResp.ID, nil)
	getW := httptest.NewRecorder()

	testAPI.Router().ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusOK, getW.Code)

	var getResp models.GetPaymentResponse
	err = json.NewDecoder(getW.Body).Decode(&getResp)
	require.NoError(t, err)

	assert.Equal(t, "Declined", getResp.Status)
}

// TestPaymentFlow_BankUnavailable tests when bank returns 503
func TestPaymentFlow_BankUnavailable(t *testing.T) {
	testAPI := api.New()
	futureYear := time.Now().Year() + 1

	reqBody := models.PostPaymentRequest{
		CardNumber:  "2222405343248870", // Ends in 0 - bank returns 503
		ExpiryMonth: 4,
		ExpiryYear:  futureYear,
		Currency:    "GBP",
		Amount:      100,
		CVV:         "123",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	testAPI.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)

	var errResp models.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Contains(t, errResp.Error, "Unable to process payment with bank")
}

// TestPaymentFlow_ValidationErrors tests various validation scenarios
func TestPaymentFlow_ValidationErrors(t *testing.T) {
	testAPI := api.New()
	futureYear := time.Now().Year() + 1

	tests := []struct {
		name          string
		request       models.PostPaymentRequest
		expectedError string
	}{
		{
			name: "invalid card number - too short",
			request: models.PostPaymentRequest{
				CardNumber:  "123",
				ExpiryMonth: 4,
				ExpiryYear:  futureYear,
				Currency:    "GBP",
				Amount:      100,
				CVV:         "123",
			},
			expectedError: "card number must be between 14-19 digits",
		},
		{
			name: "invalid card number - non-numeric",
			request: models.PostPaymentRequest{
				CardNumber:  "ABCD567890123456",
				ExpiryMonth: 4,
				ExpiryYear:  futureYear,
				Currency:    "GBP",
				Amount:      100,
				CVV:         "123",
			},
			expectedError: "card number must only contain numeric characters",
		},
		{
			name: "invalid expiry month",
			request: models.PostPaymentRequest{
				CardNumber:  "2222405343248877",
				ExpiryMonth: 13,
				ExpiryYear:  futureYear,
				Currency:    "GBP",
				Amount:      100,
				CVV:         "123",
			},
			expectedError: "expiry month must be between 1-12",
		},
		{
			name: "invalid currency",
			request: models.PostPaymentRequest{
				CardNumber:  "2222405343248877",
				ExpiryMonth: 4,
				ExpiryYear:  futureYear,
				Currency:    "JPY", // Not supported
				Amount:      100,
				CVV:         "123",
			},
			expectedError: "currency must be a valid 3-character ISO code",
		},
		{
			name: "invalid amount - zero",
			request: models.PostPaymentRequest{
				CardNumber:  "2222405343248877",
				ExpiryMonth: 4,
				ExpiryYear:  futureYear,
				Currency:    "GBP",
				Amount:      0,
				CVV:         "123",
			},
			expectedError: "amount must be a positive integer",
		},
		{
			name: "invalid CVV - too short",
			request: models.PostPaymentRequest{
				CardNumber:  "2222405343248877",
				ExpiryMonth: 4,
				ExpiryYear:  futureYear,
				Currency:    "GBP",
				Amount:      100,
				CVV:         "12",
			},
			expectedError: "CVV must be 3-4 digits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			testAPI.Router().ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var errResp models.ErrorResponse
			err := json.NewDecoder(w.Body).Decode(&errResp)
			require.NoError(t, err)

			assert.Contains(t, errResp.Error, tt.expectedError)
		})
	}
}

// TestPaymentFlow_GetNonExistent tests retrieving a non-existent payment
func TestPaymentFlow_GetNonExistent(t *testing.T) {
	testAPI := api.New()

	req := httptest.NewRequest(http.MethodGet, "/api/payments/non-existent-id", nil)
	w := httptest.NewRecorder()

	testAPI.Router().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errResp models.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errResp)
	require.NoError(t, err)

	assert.Equal(t, "Payment not found", errResp.Error)
}

// TestPaymentFlow_MultipleCurrencies tests payments with different supported currencies
func TestPaymentFlow_MultipleCurrencies(t *testing.T) {
	testAPI := api.New()
	futureYear := time.Now().Year() + 1

	currencies := []string{"USD", "GBP", "EUR"}

	for _, currency := range currencies {
		t.Run("currency_"+currency, func(t *testing.T) {
			reqBody := models.PostPaymentRequest{
				CardNumber:  "2222405343248877",
				ExpiryMonth: 4,
				ExpiryYear:  futureYear,
				Currency:    currency,
				Amount:      100,
				CVV:         "123",
			}

			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			testAPI.Router().ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var postResp models.PostPaymentResponse
			err := json.NewDecoder(w.Body).Decode(&postResp)
			require.NoError(t, err)

			assert.Equal(t, currency, postResp.Currency)
		})
	}
}

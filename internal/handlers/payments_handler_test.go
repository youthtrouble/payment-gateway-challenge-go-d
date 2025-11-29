package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) ProcessPayment(payment *domain.Payment) (*domain.Payment, error) {
	args := m.Called(payment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockPaymentService) GetPayment(id string) (*domain.Payment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func TestPostHandler_Success(t *testing.T) {
	mockService := new(MockPaymentService)
	futureYear := time.Now().Year() + 1

	processedPayment := &domain.Payment{
		ID: "generated-id-123",
		Card: domain.Card{
			Number:      "2222405343248877",
			ExpiryMonth: 12,
			ExpiryYear:  futureYear,
			CVV:         "123",
		},
		Currency: "GBP",
		Amount:   100,
		Status:   domain.StatusAuthorized,
	}

	mockService.On("ProcessPayment", mock.AnythingOfType("*domain.Payment")).Return(processedPayment, nil)

	handler := NewPaymentsHandler(mockService)

	reqBody := models.PostPaymentRequest{
		CardNumber:  "2222405343248877",
		ExpiryMonth: 12,
		ExpiryYear:  futureYear,
		Currency:    "GBP",
		Amount:      100,
		CVV:         "123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.PostHandler()(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.PostPaymentResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "generated-id-123", response.ID)
	assert.Equal(t, "Authorized", response.Status)
	assert.Equal(t, "8877", response.CardNumberLastFour)
	assert.Equal(t, 12, response.ExpiryMonth)
	assert.Equal(t, futureYear, response.ExpiryYear)
	assert.Equal(t, "GBP", response.Currency)
	assert.Equal(t, 100, response.Amount)

	mockService.AssertExpectations(t)
}

func TestPostHandler_ValidationError(t *testing.T) {
	mockService := new(MockPaymentService)
	handler := NewPaymentsHandler(mockService)

	reqBody := models.PostPaymentRequest{
		CardNumber:  "123", // Invalid - too short
		ExpiryMonth: 4,
		ExpiryYear:  2025,
		Currency:    "GBP",
		Amount:      100,
		CVV:         "123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.PostHandler()(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "card number")

	mockService.AssertNotCalled(t, "ProcessPayment")
}

func TestPostHandler_InvalidJSON(t *testing.T) {
	mockService := new(MockPaymentService)
	handler := NewPaymentsHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()

	handler.PostHandler()(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Invalid request body", response.Error)
}

func TestPostHandler_BankError(t *testing.T) {
	mockService := new(MockPaymentService)
	mockService.On("ProcessPayment", mock.AnythingOfType("*domain.Payment")).Return(nil, errors.New("bank communication error"))

	handler := NewPaymentsHandler(mockService)

	futureYear := time.Now().Year() + 1

	reqBody := models.PostPaymentRequest{
		CardNumber:  "2222405343248877",
		ExpiryMonth: 12,
		ExpiryYear:  futureYear,
		Currency:    "GBP",
		Amount:      100,
		CVV:         "123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.PostHandler()(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "Unable to process payment with bank")

	mockService.AssertExpectations(t)
}

func TestGetHandler_Success(t *testing.T) {
	mockService := new(MockPaymentService)

	expectedPayment := &domain.Payment{
		ID: "test-payment-id",
		Card: domain.Card{
			Number:      "2222405343248877",
			ExpiryMonth: 4,
			ExpiryYear:  2025,
			CVV:         "123",
		},
		Currency: "GBP",
		Amount:   100,
		Status:   domain.StatusAuthorized,
	}

	mockService.On("GetPayment", "test-payment-id").Return(expectedPayment, nil)

	handler := NewPaymentsHandler(mockService)

	r := chi.NewRouter()
	r.Get("/api/payments/{id}", handler.GetHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/payments/test-payment-id", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.GetPaymentResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "test-payment-id", response.ID)
	assert.Equal(t, "Authorized", response.Status)
	assert.Equal(t, "8877", response.CardNumberLastFour)

	mockService.AssertExpectations(t)
}

func TestGetHandler_NotFound(t *testing.T) {
	mockService := new(MockPaymentService)
	mockService.On("GetPayment", "non-existent-id").Return(nil, domain.ErrPaymentNotFound)

	handler := NewPaymentsHandler(mockService)

	r := chi.NewRouter()
	r.Get("/api/payments/{id}", handler.GetHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/payments/non-existent-id", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Payment not found", response.Error)

	mockService.AssertExpectations(t)
}

func TestGetHandler_InternalError(t *testing.T) {
	mockService := new(MockPaymentService)
	mockService.On("GetPayment", "test-id").Return(nil, errors.New("database error"))

	handler := NewPaymentsHandler(mockService)

	r := chi.NewRouter()
	r.Get("/api/payments/{id}", handler.GetHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/payments/test-id", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Failed to retrieve payment", response.Error)

	mockService.AssertExpectations(t)
}

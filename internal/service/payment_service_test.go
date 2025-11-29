package service

import (
	"errors"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/client"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockBankClient struct {
	mock.Mock
}

func (m *MockBankClient) ProcessPayment(payment *domain.Payment) (*client.BankResponse, error) {
	args := m.Called(payment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.BankResponse), args.Error(1)
}

type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Save(payment *domain.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) FindByID(id string) (*domain.Payment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func TestPaymentService_ProcessPayment_Authorized(t *testing.T) {

	mockBank := new(MockBankClient)
	mockRepo := new(MockPaymentRepository)

	payment := &domain.Payment{
		Card: domain.Card{
			Number:      "2222405343248877", // Odd ending - authorized
			ExpiryMonth: 4,
			ExpiryYear:  2025,
			CVV:         "123",
		},
		Currency: "GBP",
		Amount:   100,
		Status:   domain.StatusRejected,
	}

	mockBank.On("ProcessPayment", payment).Return(&client.BankResponse{
		Authorized:        true,
		AuthorizationCode: "auth-code-123",
	}, nil)

	mockRepo.On("Save", payment).Return(nil)

	service := NewPaymentService(mockBank, mockRepo)

	result, err := service.ProcessPayment(payment)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ID) // Should have generated an ID
	assert.Equal(t, domain.StatusAuthorized, result.Status)

	mockBank.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestPaymentService_ProcessPayment_Declined(t *testing.T) {

	mockBank := new(MockBankClient)
	mockRepo := new(MockPaymentRepository)

	payment := &domain.Payment{
		Card: domain.Card{
			Number:      "2222405343248878", // Even ending - declined
			ExpiryMonth: 4,
			ExpiryYear:  2025,
			CVV:         "123",
		},
		Currency: "GBP",
		Amount:   100,
		Status:   domain.StatusRejected,
	}

	mockBank.On("ProcessPayment", payment).Return(&client.BankResponse{
		Authorized:        false,
		AuthorizationCode: "",
	}, nil)

	mockRepo.On("Save", payment).Return(nil)

	service := NewPaymentService(mockBank, mockRepo)

	result, err := service.ProcessPayment(payment)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, domain.StatusDeclined, result.Status)

	mockBank.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestPaymentService_ProcessPayment_BankError(t *testing.T) {

	mockBank := new(MockBankClient)
	mockRepo := new(MockPaymentRepository)

	payment := &domain.Payment{
		Card: domain.Card{
			Number:      "2222405343248870", // Ends in 0 - bank error
			ExpiryMonth: 4,
			ExpiryYear:  2025,
			CVV:         "123",
		},
		Currency: "GBP",
		Amount:   100,
	}

	mockBank.On("ProcessPayment", payment).Return(nil, errors.New("bank service unavailable"))

	// Repository should NOT be called since bank failed

	service := NewPaymentService(mockBank, mockRepo)

	result, err := service.ProcessPayment(payment)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to process payment with bank")

	mockBank.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Save") // Should not save if bank fails
}

func TestPaymentService_ProcessPayment_RepositoryError(t *testing.T) {

	mockBank := new(MockBankClient)
	mockRepo := new(MockPaymentRepository)

	payment := &domain.Payment{
		Card: domain.Card{
			Number:      "2222405343248877",
			ExpiryMonth: 4,
			ExpiryYear:  2025,
			CVV:         "123",
		},
		Currency: "GBP",
		Amount:   100,
	}

	mockBank.On("ProcessPayment", payment).Return(&client.BankResponse{
		Authorized:        true,
		AuthorizationCode: "auth-code-123",
	}, nil)

	mockRepo.On("Save", payment).Return(errors.New("database error"))

	service := NewPaymentService(mockBank, mockRepo)

	result, err := service.ProcessPayment(payment)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save payment")

	mockBank.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestPaymentService_GetPayment_Found(t *testing.T) {

	mockBank := new(MockBankClient)
	mockRepo := new(MockPaymentRepository)

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

	mockRepo.On("FindByID", "test-payment-id").Return(expectedPayment, nil)

	service := NewPaymentService(mockBank, mockRepo)

	result, err := service.GetPayment("test-payment-id")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedPayment.ID, result.ID)
	assert.Equal(t, expectedPayment.Status, result.Status)

	mockRepo.AssertExpectations(t)
}

func TestPaymentService_GetPayment_NotFound(t *testing.T) {

	mockBank := new(MockBankClient)
	mockRepo := new(MockPaymentRepository)

	mockRepo.On("FindByID", "non-existent-id").Return(nil, nil)

	service := NewPaymentService(mockBank, mockRepo)

	result, err := service.GetPayment("non-existent-id")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrPaymentNotFound, err)

	mockRepo.AssertExpectations(t)
}

func TestPaymentService_GetPayment_RepositoryError(t *testing.T) {

	mockBank := new(MockBankClient)
	mockRepo := new(MockPaymentRepository)
	mockRepo.On("FindByID", "test-id").Return(nil, errors.New("database connection error"))

	service := NewPaymentService(mockBank, mockRepo)

	result, err := service.GetPayment("test-id")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database connection error")

	mockRepo.AssertExpectations(t)
}

func TestPaymentService_IDGeneration(t *testing.T) {

	mockBank := new(MockBankClient)
	mockRepo := new(MockPaymentRepository)
	mockBank.On("ProcessPayment", mock.Anything).Return(&client.BankResponse{
		Authorized:        true,
		AuthorizationCode: "auth-code",
	}, nil)

	mockRepo.On("Save", mock.Anything).Return(nil)

	service := NewPaymentService(mockBank, mockRepo)

	payment1 := &domain.Payment{
		Card: domain.Card{
			Number:      "1234567890123456",
			ExpiryMonth: 12,
			ExpiryYear:  2025,
			CVV:         "123",
		},
		Currency: "USD",
		Amount:   100,
	}

	payment2 := &domain.Payment{
		Card: domain.Card{
			Number:      "1234567890123456",
			ExpiryMonth: 12,
			ExpiryYear:  2025,
			CVV:         "123",
		},
		Currency: "USD",
		Amount:   100,
	}

	result1, _ := service.ProcessPayment(payment1)
	result2, _ := service.ProcessPayment(payment2)

	assert.NotEmpty(t, result1.ID)
	assert.NotEmpty(t, result2.ID)
	assert.NotEqual(t, result1.ID, result2.ID)
}

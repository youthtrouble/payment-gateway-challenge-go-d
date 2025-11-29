package service

import (
	"fmt"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/client"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain"
	"github.com/google/uuid"
)

type PaymentRepository interface {
	Save(payment *domain.Payment) error
	FindByID(id string) (*domain.Payment, error)
}

type PaymentService struct {
	bankClient client.BankClient
	repository PaymentRepository
}

func NewPaymentService(bankClient client.BankClient, repository PaymentRepository) *PaymentService {
	return &PaymentService{
		bankClient: bankClient,
		repository: repository,
	}
}

// 1. Validate the payment (already done in domain)
// 2. Call the bank to authorize
// 3. Update payment status based on bank response
// 4. Store the payment
// 5. Return the payment
func (s *PaymentService) ProcessPayment(payment *domain.Payment) (*domain.Payment, error) {
	payment.ID = uuid.New().String()

	bankResp, err := s.bankClient.ProcessPayment(payment)
	if err != nil {
		// If bank is unavailable or returns an error, we don't store the payment
		// This is a rejection at the bank level
		return nil, fmt.Errorf("failed to process payment with bank: %w", err)
	}

	if bankResp.Authorized {
		payment.SetAuthorized()
	} else {
		payment.SetDeclined()
	}

	if err := s.repository.Save(payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	return payment, nil
}

func (s *PaymentService) GetPayment(id string) (*domain.Payment, error) {
	payment, err := s.repository.FindByID(id)
	if err != nil {
		return nil, err
	}

	if payment == nil {
		return nil, domain.ErrPaymentNotFound
	}

	return payment, nil
}

package repository

import (
	"sync"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain"
)

// In production, this would be replaced with a database implementation
type PaymentsRepository struct {
	payments map[string]*domain.Payment
	mu       sync.RWMutex // Thread-safe for concurrent access
}

func NewPaymentsRepository() *PaymentsRepository {
	return &PaymentsRepository{
		payments: make(map[string]*domain.Payment),
	}
}

func (r *PaymentsRepository) Save(payment *domain.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.payments[payment.ID] = payment
	return nil
}

func (r *PaymentsRepository) FindByID(id string) (*domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	payment, exists := r.payments[id]
	if !exists {
		return nil, nil
	}

	return payment, nil
}

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/go-chi/chi/v5"
)

type PaymentService interface {
	ProcessPayment(payment *domain.Payment) (*domain.Payment, error)
	GetPayment(id string) (*domain.Payment, error)
}

type PaymentsHandler struct {
	paymentService PaymentService
}

func NewPaymentsHandler(paymentService PaymentService) *PaymentsHandler {
	return &PaymentsHandler{
		paymentService: paymentService,
	}
}

func (h *PaymentsHandler) PostHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req models.PostPaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		payment, err := req.ToDomainPayment()
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		processedPayment, err := h.paymentService.ProcessPayment(payment)
		if err != nil {

			h.respondWithError(w, http.StatusBadGateway, "Unable to process payment with bank")
			return
		}

		response := models.FromDomainPayment(processedPayment)

		h.respondWithJSON(w, http.StatusOK, response)
	}
}

func (h *PaymentsHandler) GetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := chi.URLParam(r, "id")
		if id == "" {
			h.respondWithError(w, http.StatusBadRequest, "Payment ID is required")
			return
		}

		payment, err := h.paymentService.GetPayment(id)
		if err != nil {
			if errors.Is(err, domain.ErrPaymentNotFound) {
				h.respondWithError(w, http.StatusNotFound, "Payment not found")
				return
			}

			h.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve payment")
			return
		}

		response := models.ToGetPaymentResponse(payment)

		h.respondWithJSON(w, http.StatusOK, response)
	}
}

func (h *PaymentsHandler) respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *PaymentsHandler) respondWithError(w http.ResponseWriter, statusCode int, message string) {
	h.respondWithJSON(w, statusCode, models.ErrorResponse{Error: message})
}

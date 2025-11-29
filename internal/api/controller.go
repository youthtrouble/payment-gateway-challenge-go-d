package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cko-recruitment/payment-gateway-challenge-go/docs"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/handlers"
	httpSwagger "github.com/swaggo/http-swagger"
)

type pong struct {
	Message string `json:"message"`
}

// PingHandler returns an http.HandlerFunc that handles HTTP Ping GET requests.
func (a *Api) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(pong{Message: "pong"}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// SwaggerHandler returns an http.HandlerFunc that handles HTTP Swagger related requests.
func (a *Api) SwaggerHandler() http.HandlerFunc {
	return httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s/swagger/doc.json", docs.SwaggerInfo.Host)),
	)
}

// PostPaymentHandler godoc
// @Summary Process a new payment
// @Description Process a payment through the payment gateway and return the result
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body models.PostPaymentRequest true "Payment details"
// @Success 200 {object} models.PostPaymentResponse "Payment processed successfully (Authorized or Declined)"
// @Failure 400 {object} models.ErrorResponse "Invalid request or validation error (Rejected)"
// @Failure 502 {object} models.ErrorResponse "Bank service unavailable or error"
// @Router /api/payments [post]
func (a *Api) PostPaymentHandler() http.HandlerFunc {
	h := handlers.NewPaymentsHandler(a.paymentService)
	return h.PostHandler()
}

// GetPaymentHandler godoc
// @Summary Retrieve a payment by ID
// @Description Get details of a previously processed payment
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} models.GetPaymentResponse "Payment found"
// @Failure 404 {object} models.ErrorResponse "Payment not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/payments/{id} [get]
func (a *Api) GetPaymentHandler() http.HandlerFunc {
	h := handlers.NewPaymentsHandler(a.paymentService)
	return h.GetHandler()
}

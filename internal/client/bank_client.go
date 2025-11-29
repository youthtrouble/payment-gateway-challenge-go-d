package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/domain"
)

// BankClient defines the interface for communicating with the acquiring bank
type BankClient interface {
	ProcessPayment(payment *domain.Payment) (*BankResponse, error)
}

// BankRequest represents the request format expected by the bank simulator
type BankRequest struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	Currency   string `json:"currency"`
	Amount     int    `json:"amount"`
	CVV        string `json:"cvv"`
}

// BankResponse represents the response from the bank simulator
type BankResponse struct {
	Authorized        bool   `json:"authorized"`
	AuthorizationCode string `json:"authorization_code"`
}

// HTTPBankClient is an HTTP implementation of BankClient
type HTTPBankClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPBankClient creates a new HTTP bank client
func NewHTTPBankClient(baseURL string) *HTTPBankClient {
	return &HTTPBankClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *HTTPBankClient) ProcessPayment(payment *domain.Payment) (*BankResponse, error) {
	bankReq := c.convertTobankRequest(payment)

	jsonData, err := json.Marshal(bankReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bank request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/payments", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to bank: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var bankResp BankResponse
		if err := json.Unmarshal(body, &bankResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal bank response: %w", err)
		}
		return &bankResp, nil

	case http.StatusBadRequest:
		return nil, fmt.Errorf("bank rejected request: %s", string(body))

	case http.StatusServiceUnavailable:
		return nil, fmt.Errorf("bank service unavailable")

	default:
		return nil, fmt.Errorf("unexpected response from bank: %d - %s", resp.StatusCode, string(body))
	}
}

func (c *HTTPBankClient) convertTobankRequest(payment *domain.Payment) *BankRequest {
	// Format expiry date as MM/YYYY
	expiryDate := fmt.Sprintf("%02d/%d", payment.Card.ExpiryMonth, payment.Card.ExpiryYear)

	return &BankRequest{
		CardNumber: payment.Card.Number,
		ExpiryDate: expiryDate,
		Currency:   payment.Currency,
		Amount:     payment.Amount,
		CVV:        payment.Card.CVV,
	}
}

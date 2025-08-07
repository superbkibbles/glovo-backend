package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"glovo-backend/services/order-service/internal/domain"

	"github.com/google/uuid"
)

type paymentClient struct {
	baseURL string
	client  *http.Client
}

func NewPaymentClient() domain.PaymentService {
	baseURL := getEnv("PAYMENT_SERVICE_URL", "http://localhost:8007")
	return &paymentClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (p *paymentClient) ProcessPayment(orderID string, amount float64, paymentInfo domain.PaymentInfo) (*domain.PaymentResult, error) {
	url := fmt.Sprintf("%s/api/v1/payments/process", p.baseURL)

	reqBody := map[string]interface{}{
		"order_id":  orderID,
		"amount":    amount,
		"method":    paymentInfo.Method,
		"reference": paymentInfo.Reference,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment request: %w", err)
	}

	resp, err := p.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to process payment: %w", err)
	}
	defer resp.Body.Close()

	var result domain.PaymentResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode payment response: %w", err)
	}

	return &result, nil
}

// Mock implementation for development
type mockPaymentClient struct{}

func NewMockPaymentClient() domain.PaymentService {
	return &mockPaymentClient{}
}

func (m *mockPaymentClient) ProcessPayment(orderID string, amount float64, paymentInfo domain.PaymentInfo) (*domain.PaymentResult, error) {
	log.Printf("MOCK: Processing payment for order %s, amount: $%.2f, method: %s", orderID, amount, paymentInfo.Method)

	// Simulate payment processing
	reference := uuid.New().String()

	// Simulate payment success (90% success rate for testing)
	success := true // For demo purposes, always succeed

	result := &domain.PaymentResult{
		Success:   success,
		Reference: reference,
	}

	if !success {
		result.Error = "Payment processing failed (mock error)"
	}

	return result, nil
}

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"glovo-backend/services/order-service/internal/domain"
)

type catalogClient struct {
	baseURL string
	client  *http.Client
}

func NewCatalogClient() domain.CatalogService {
	baseURL := getEnv("CATALOG_SERVICE_URL", "http://localhost:8003")
	return &catalogClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *catalogClient) GetProduct(productID string) (*domain.Product, error) {
	url := fmt.Sprintf("%s/api/v1/products/%s", c.baseURL, productID)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("product service returned status %d", resp.StatusCode)
	}

	var product domain.Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		return nil, fmt.Errorf("failed to decode product response: %w", err)
	}

	return &product, nil
}

func (c *catalogClient) ValidateOrder(merchantID string, items []domain.OrderItemReq) (*domain.OrderValidation, error) {
	url := fmt.Sprintf("%s/api/v1/stores/%s/validate-order", c.baseURL, merchantID)

	reqBody := map[string]interface{}{
		"items": items,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to validate order: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalog service returned status %d", resp.StatusCode)
	}

	var validation domain.OrderValidation
	if err := json.NewDecoder(resp.Body).Decode(&validation); err != nil {
		return nil, fmt.Errorf("failed to decode validation response: %w", err)
	}

	return &validation, nil
}

// Mock implementation for development
type mockCatalogClient struct{}

func NewMockCatalogClient() domain.CatalogService {
	return &mockCatalogClient{}
}

func (m *mockCatalogClient) GetProduct(productID string) (*domain.Product, error) {
	// Mock product data
	return &domain.Product{
		ID:          productID,
		Name:        "Mock Product",
		Price:       12.99,
		Description: "Mock product for testing",
		Available:   true,
	}, nil
}

func (m *mockCatalogClient) ValidateOrder(merchantID string, items []domain.OrderItemReq) (*domain.OrderValidation, error) {
	var validatedItems []domain.ValidatedItem
	var totalAmount float64

	for _, item := range items {
		validatedItem := domain.ValidatedItem{
			ProductID: item.ProductID,
			Name:      fmt.Sprintf("Product %s", item.ProductID),
			Price:     12.99,
			Quantity:  item.Quantity,
			Available: true,
			Subtotal:  12.99 * float64(item.Quantity),
		}
		validatedItems = append(validatedItems, validatedItem)
		totalAmount += validatedItem.Subtotal
	}

	return &domain.OrderValidation{
		Valid:       true,
		Items:       validatedItems,
		TotalAmount: totalAmount,
		Errors:      []string{},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

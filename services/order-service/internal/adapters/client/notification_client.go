package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"glovo-backend/services/order-service/internal/domain"
)

type notificationClient struct {
	baseURL string
	client  *http.Client
}

func NewNotificationClient() domain.NotificationService {
	baseURL := getEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8008")
	return &notificationClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (n *notificationClient) SendOrderNotification(orderID string, userID string, message string) error {
	url := fmt.Sprintf("%s/api/v1/notifications/send", n.baseURL)

	reqBody := map[string]interface{}{
		"user_id": userID,
		"type":    "order_update",
		"title":   "Order Update",
		"message": message,
		"metadata": map[string]string{
			"order_id": orderID,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal notification request: %w", err)
	}

	resp, err := n.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}

	return nil
}

// Mock implementation for development
type mockNotificationClient struct{}

func NewMockNotificationClient() domain.NotificationService {
	return &mockNotificationClient{}
}

func (m *mockNotificationClient) SendOrderNotification(orderID string, userID string, message string) error {
	log.Printf("MOCK NOTIFICATION - Order: %s, User: %s, Message: %s", orderID, userID, message)
	return nil
}

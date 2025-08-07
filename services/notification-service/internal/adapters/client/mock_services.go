package client

import (
	"fmt"
	"log"
)

// Mock Push Notification Service
type mockPushNotificationService struct{}

func NewMockPushNotificationService() *mockPushNotificationService {
	return &mockPushNotificationService{}
}

func (m *mockPushNotificationService) SendPushNotification(deviceTokens []string, title, message string, data map[string]string) error {
	log.Printf("Mock Push Notification sent to %d devices - Title: %s, Message: %s", len(deviceTokens), title, message)
	return nil
}

func (m *mockPushNotificationService) SendPushToTopic(topic, title, message string, data map[string]string) error {
	log.Printf("Mock Push Notification sent to topic %s - Title: %s, Message: %s", topic, title, message)
	return nil
}

// Mock SMS Service
type mockSMSService struct{}

func NewMockSMSService() *mockSMSService {
	return &mockSMSService{}
}

func (m *mockSMSService) SendSMS(phoneNumber, message string) error {
	log.Printf("Mock SMS sent to %s: %s", phoneNumber, message)
	return nil
}

func (m *mockSMSService) SendBulkSMS(phoneNumbers []string, message string) error {
	log.Printf("Mock Bulk SMS sent to %d numbers: %s", len(phoneNumbers), message)
	return nil
}

// Mock Email Service
type mockEmailService struct{}

func NewMockEmailService() *mockEmailService {
	return &mockEmailService{}
}

func (m *mockEmailService) SendEmail(to, subject, body string) error {
	log.Printf("Mock Email sent to %s - Subject: %s, Body: %s", to, subject, body)
	return nil
}

func (m *mockEmailService) SendBulkEmail(recipients []string, subject, body string) error {
	log.Printf("Mock Bulk Email sent to %d recipients - Subject: %s", len(recipients), subject)
	return nil
}

func (m *mockEmailService) SendTemplateEmail(to, templateID string, variables map[string]string) error {
	log.Printf("Mock Template Email sent to %s using template %s with variables: %v", to, templateID, variables)
	return nil
}

// Twilio SMS Service (Real implementation placeholder)
type twilioSMSService struct {
	accountSID string
	authToken  string
	fromNumber string
}

func NewTwilioSMSService(accountSID, authToken, fromNumber string) *twilioSMSService {
	return &twilioSMSService{
		accountSID: accountSID,
		authToken:  authToken,
		fromNumber: fromNumber,
	}
}

func (t *twilioSMSService) SendSMS(phoneNumber, message string) error {
	// In production, this would integrate with actual Twilio API
	log.Printf("Twilio SMS to %s: %s", phoneNumber, message)
	return nil
}

func (t *twilioSMSService) SendBulkSMS(phoneNumbers []string, message string) error {
	for _, phone := range phoneNumbers {
		if err := t.SendSMS(phone, message); err != nil {
			return fmt.Errorf("failed to send SMS to %s: %w", phone, err)
		}
	}
	return nil
}

// Firebase Push Notification Service (Real implementation placeholder)
type firebasePushService struct {
	serverKey string
}

func NewFirebasePushService(serverKey string) *firebasePushService {
	return &firebasePushService{
		serverKey: serverKey,
	}
}

func (f *firebasePushService) SendPushNotification(deviceTokens []string, title, message string, data map[string]string) error {
	// In production, this would integrate with Firebase Cloud Messaging
	log.Printf("Firebase Push to %d devices - Title: %s, Message: %s", len(deviceTokens), title, message)
	return nil
}

func (f *firebasePushService) SendPushToTopic(topic, title, message string, data map[string]string) error {
	// In production, this would send to a Firebase topic
	log.Printf("Firebase Push to topic %s - Title: %s, Message: %s", topic, title, message)
	return nil
}

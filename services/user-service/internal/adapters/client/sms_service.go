package client

import (
	"fmt"
	"log"
	"os"

	"glovo-backend/services/user-service/internal/domain"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type smsService struct {
	client *twilio.RestClient
	from   string
}

func NewSMSService() domain.SMSService {
	accountSid := getEnv("TWILIO_ACCOUNT_SID", "")
	authToken := getEnv("TWILIO_AUTH_TOKEN", "")
	from := getEnv("TWILIO_PHONE_NUMBER", "")

	if accountSid == "" || authToken == "" {
		log.Println("Warning: Twilio credentials not configured, SMS will be mocked")
		return &mockSMSService{}
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	return &smsService{
		client: client,
		from:   from,
	}
}

func (s *smsService) SendSMS(phoneNumber, message string) error {
	params := &twilioApi.CreateMessageParams{}
	params.SetTo(phoneNumber)
	params.SetFrom(s.from)
	params.SetBody(message)

	_, err := s.client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	log.Printf("SMS sent to %s: %s", phoneNumber, message)
	return nil
}

// Mock SMS service for development/testing
type mockSMSService struct{}

func (m *mockSMSService) SendSMS(phoneNumber, message string) error {
	log.Printf("MOCK SMS to %s: %s", phoneNumber, message)
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

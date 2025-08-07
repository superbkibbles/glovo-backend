package client

import (
	"time"

	"glovo-backend/services/driver-service/internal/domain"
)

// Mock User Service
type mockUserService struct{}

func NewMockUserService() domain.UserService {
	return &mockUserService{}
}

func (m *mockUserService) GetUser(userID string) (*domain.User, error) {
	return &domain.User{
		ID:    userID,
		Phone: "+1234567890",
		Email: "user@example.com",
		Role:  "customer",
	}, nil
}

func (m *mockUserService) ValidateDriverRole(userID string) error {
	return nil
}

// Mock Location Service
type mockLocationService struct{}

func NewMockLocationService() domain.LocationService {
	return &mockLocationService{}
}

func (m *mockLocationService) UpdateDriverLocation(driverID string, latitude, longitude float64) error {
	return nil
}

func (m *mockLocationService) GetDriverLocation(driverID string) (*domain.CurrentLocation, error) {
	return &domain.CurrentLocation{
		Latitude:  40.7128,
		Longitude: -74.0060,
		UpdatedAt: time.Now(),
	}, nil
}

// Mock Payment Service
type mockPaymentService struct{}

func NewMockPaymentService() domain.PaymentService {
	return &mockPaymentService{}
}

func (m *mockPaymentService) ProcessDriverPayout(driverID string, amount float64) error {
	return nil
}

func (m *mockPaymentService) GetDriverEarnings(driverID string, startDate, endDate time.Time) (*domain.EarningsReport, error) {
	return &domain.EarningsReport{
		Period:      "Weekly",
		Deliveries:  25,
		Earnings:    750.50,
		Commission:  112.58,
		NetEarnings: 637.92,
	}, nil
}

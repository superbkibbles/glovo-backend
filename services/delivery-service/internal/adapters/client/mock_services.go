package client

import (
	"glovo-backend/services/delivery-service/internal/domain"
)

// Mock Order Service
type mockOrderService struct{}

func NewMockOrderService() domain.OrderService {
	return &mockOrderService{}
}

func (m *mockOrderService) GetOrder(orderID string) (*domain.OrderInfo, error) {
	return &domain.OrderInfo{
		ID:           orderID,
		CustomerName: "John Doe",
		Items:        3,
		TotalAmount:  29.99,
	}, nil
}

func (m *mockOrderService) UpdateOrderStatus(orderID string, status string) error {
	return nil
}

// Mock Driver Service
type mockDriverService struct{}

func NewMockDriverService() domain.DriverService {
	return &mockDriverService{}
}

func (m *mockDriverService) GetDriver(driverID string) (*domain.DriverInfo, error) {
	return &domain.DriverInfo{
		ID:     driverID,
		Name:   "Jane Driver",
		Phone:  "+1234567890",
		Rating: 4.8,
		Vehicle: domain.Vehicle{
			Type:         "motorcycle",
			Make:         "Honda",
			Model:        "CBR600",
			LicensePlate: "ABC123",
		},
	}, nil
}

func (m *mockDriverService) GetAvailableDrivers(latitude, longitude, radius float64) ([]domain.DriverAvailability, error) {
	return []domain.DriverAvailability{
		{
			DriverID:    "driver1",
			Name:        "John Driver",
			Distance:    2.5,
			Rating:      4.8,
			ETA:         10,
			IsOnline:    true,
			IsAvailable: true,
		},
	}, nil
}

func (m *mockDriverService) IsDriverAvailable(driverID string) (bool, error) {
	return true, nil
}

func (m *mockDriverService) UpdateDriverStatus(driverID string, status string) error {
	return nil
}

// Mock Location Service
type mockLocationService struct{}

func NewMockLocationService() domain.LocationService {
	return &mockLocationService{}
}

func (m *mockLocationService) GetDriverLocation(driverID string) (*domain.Location, error) {
	return &domain.Location{
		Latitude:  40.7128,
		Longitude: -74.0060,
	}, nil
}

func (m *mockLocationService) CreateDeliveryRoute(deliveryID, driverID string, pickup, dropoff domain.Location) error {
	return nil
}

func (m *mockLocationService) GetDeliveryTracking(deliveryID string) (*domain.TrackingInfo, error) {
	return &domain.TrackingInfo{
		CurrentLocation: &domain.Location{
			Latitude:  40.7128,
			Longitude: -74.0060,
		},
	}, nil
}

func (m *mockLocationService) CalculateETA(from, to domain.Location) (int, error) {
	return 15, nil
}

// Mock Notification Service
type mockNotificationService struct{}

func NewMockNotificationService() domain.NotificationService {
	return &mockNotificationService{}
}

func (m *mockNotificationService) SendDeliveryAssignment(driverID string, delivery *domain.Delivery) error {
	return nil
}

func (m *mockNotificationService) SendDeliveryUpdate(orderID string, status domain.DeliveryStatus) error {
	return nil
}

func (m *mockNotificationService) SendDriverNotification(driverID string, message string) error {
	return nil
}

// Mock Payment Service
type mockPaymentService struct{}

func NewMockPaymentService() domain.PaymentService {
	return &mockPaymentService{}
}

func (m *mockPaymentService) ProcessDeliveryPayment(deliveryID string) error {
	return nil
}

func (m *mockPaymentService) CalculateDriverPayout(deliveryID string) (float64, error) {
	return 15.50, nil
}

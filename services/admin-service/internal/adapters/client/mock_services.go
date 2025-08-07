package client

import (
	"log"
	"time"

	"glovo-backend/services/admin-service/internal/domain"
	"glovo-backend/shared/auth"
)

// Mock UserService
type mockUserService struct{}

func NewMockUserService() domain.UserService {
	return &mockUserService{}
}

func (m *mockUserService) GetUser(userID string) (*domain.UserInfo, error) {
	return &domain.UserInfo{
		ID:    userID,
		Phone: "+1234567890",
		Email: "user@example.com",
		Role:  auth.RoleCustomer,
		Profile: domain.UserProfile{
			FirstName: "John",
			LastName:  "Doe",
		},
		Status:    domain.UserStatusActive,
		CreatedAt: time.Now().AddDate(0, -1, 0),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockUserService) GetUsers(req domain.UserSearchRequest) ([]domain.UserInfo, error) {
	return []domain.UserInfo{
		{
			ID:    "user-1",
			Phone: "+1234567890",
			Role:  auth.RoleCustomer,
			Profile: domain.UserProfile{
				FirstName: "John",
				LastName:  "Doe",
			},
			Status:    domain.UserStatusActive,
			CreatedAt: time.Now().AddDate(0, -1, 0),
		},
		{
			ID:    "user-2",
			Phone: "+1234567891",
			Role:  auth.RoleMerchant,
			Profile: domain.UserProfile{
				FirstName: "Jane",
				LastName:  "Smith",
			},
			Status:    domain.UserStatusActive,
			CreatedAt: time.Now().AddDate(0, -2, 0),
		},
	}, nil
}

func (m *mockUserService) UpdateUserStatus(userID string, status domain.UserStatus, reason string) error {
	log.Printf("Mock: Updated user %s status to %s, reason: %s", userID, status, reason)
	return nil
}

func (m *mockUserService) GetMerchants(limit, offset int) ([]domain.MerchantInfo, error) {
	return []domain.MerchantInfo{
		{
			ID:           "merchant-1",
			BusinessName: "Pizza Palace",
			Profile: domain.MerchantProfile{
				BusinessName: "Pizza Palace LLC",
				Address:      "123 Main St",
				Phone:        "+1234567890",
			},
			Status:    domain.MerchantStatusActive,
			CreatedAt: time.Now().AddDate(0, -3, 0),
		},
	}, nil
}

func (m *mockUserService) GetDrivers(limit, offset int) ([]domain.DriverInfo, error) {
	return []domain.DriverInfo{
		{
			ID: "driver-1",
			Profile: domain.DriverProfile{
				FirstName: "Mike",
				LastName:  "Wilson",
				Phone:     "+1234567892",
			},
			Status:    domain.DriverStatusActive,
			CreatedAt: time.Now().AddDate(0, -2, 0),
		},
	}, nil
}

// Mock OrderService
type mockOrderService struct{}

func NewMockOrderService() domain.OrderService {
	return &mockOrderService{}
}

func (m *mockOrderService) GetOrderStats(startDate, endDate time.Time) (map[string]int, error) {
	return map[string]int{
		"pending":   45,
		"confirmed": 120,
		"preparing": 35,
		"ready":     28,
		"picked_up": 15,
		"delivered": 890,
		"cancelled": 12,
	}, nil
}

func (m *mockOrderService) GetActiveOrdersCount() (int, error) {
	return 243, nil
}

func (m *mockOrderService) GetTotalOrdersCount() (int, error) {
	return 12547, nil
}

// Mock PaymentService
type mockPaymentService struct{}

func NewMockPaymentService() domain.PaymentService {
	return &mockPaymentService{}
}

func (m *mockPaymentService) GetRevenueStats(startDate, endDate time.Time) ([]domain.RevenueStat, error) {
	return []domain.RevenueStat{
		{
			Period:  "2024-01",
			Revenue: 45780.50,
			Orders:  1243,
		},
		{
			Period:  "2024-02",
			Revenue: 52340.75,
			Orders:  1456,
		},
		{
			Period:  "2024-03",
			Revenue: 48960.25,
			Orders:  1387,
		},
	}, nil
}

func (m *mockPaymentService) GetTotalRevenue() (float64, error) {
	return 487562.45, nil
}

func (m *mockPaymentService) GetAverageOrderValue() (float64, error) {
	return 38.75, nil
}

// Mock CatalogService
type mockCatalogService struct{}

func NewMockCatalogService() domain.CatalogService {
	return &mockCatalogService{}
}

func (m *mockCatalogService) GetTopMerchants(limit int) ([]domain.MerchantStat, error) {
	return []domain.MerchantStat{
		{
			ID:      "merchant-1",
			Name:    "Pizza Palace",
			Orders:  456,
			Revenue: 18450.75,
			Rating:  4.8,
		},
		{
			ID:      "merchant-2",
			Name:    "Burger House",
			Orders:  389,
			Revenue: 15670.25,
			Rating:  4.6,
		},
		{
			ID:      "merchant-3",
			Name:    "Sushi Express",
			Orders:  234,
			Revenue: 12890.50,
			Rating:  4.9,
		},
	}, nil
}

// Mock DriverService
type mockDriverService struct{}

func NewMockDriverService() domain.DriverService {
	return &mockDriverService{}
}

func (m *mockDriverService) GetTopDrivers(limit int) ([]domain.DriverStat, error) {
	return []domain.DriverStat{
		{
			ID:         "driver-1",
			Name:       "Mike Wilson",
			Deliveries: 245,
			Rating:     4.9,
			Earnings:   4890.75,
		},
		{
			ID:         "driver-2",
			Name:       "Sarah Davis",
			Deliveries: 198,
			Rating:     4.8,
			Earnings:   3950.50,
		},
		{
			ID:         "driver-3",
			Name:       "Tom Johnson",
			Deliveries: 167,
			Rating:     4.7,
			Earnings:   3340.25,
		},
	}, nil
}

// Mock AnalyticsService
type mockAnalyticsService struct{}

func NewMockAnalyticsService() domain.AnalyticsService {
	return &mockAnalyticsService{}
}

func (m *mockAnalyticsService) GetPlatformStats() (*domain.PlatformStats, error) {
	return &domain.PlatformStats{
		TotalUsers:        15420,
		TotalMerchants:    1250,
		TotalDrivers:      890,
		ActiveOrders:      243,
		TotalOrders:       12547,
		TotalRevenue:      487562.45,
		MonthlyRevenue:    48960.25,
		DailyRevenue:      1632.00,
		AverageOrderValue: 38.75,
		TopMerchants: []domain.MerchantStat{
			{ID: "merchant-1", Name: "Pizza Palace", Orders: 456, Revenue: 18450.75, Rating: 4.8},
		},
		TopDrivers: []domain.DriverStat{
			{ID: "driver-1", Name: "Mike Wilson", Deliveries: 245, Rating: 4.9, Earnings: 4890.75},
		},
		OrdersByStatus: map[string]int{
			"pending":   45,
			"confirmed": 120,
			"delivered": 890,
		},
		RevenueByPeriod: []domain.RevenueStat{
			{Period: "2024-01", Revenue: 45780.50, Orders: 1243},
			{Period: "2024-02", Revenue: 52340.75, Orders: 1456},
		},
	}, nil
}

func (m *mockAnalyticsService) GetRevenueAnalytics(startDate, endDate time.Time) ([]domain.RevenueStat, error) {
	days := int(endDate.Sub(startDate).Hours() / 24)
	stats := make([]domain.RevenueStat, 0, days)

	for i := 0; i < days && i < 30; i++ {
		date := startDate.AddDate(0, 0, i)
		stats = append(stats, domain.RevenueStat{
			Period:  date.Format("2006-01-02"),
			Revenue: float64(1200 + i*50),
			Orders:  30 + i*2,
		})
	}

	return stats, nil
}

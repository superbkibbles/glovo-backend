package main

import (
	"log"
	"net/http"
	"os"
	"time"

	httpHandler "glovo-backend/services/analytics-service/internal/adapters/http"
	"glovo-backend/services/analytics-service/internal/app"
	"glovo-backend/services/analytics-service/internal/domain"
	"glovo-backend/shared/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Analytics Service API
// @version 1.0
// @description Platform analytics and business intelligence service
// @host localhost:8010
// @BasePath /

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize database connection
	postgresDB := database.ConnectPostgres()

	// Auto-migrate database schema
	if err := postgresDB.AutoMigrate(
		&domain.PlatformMetrics{},
		&domain.RevenueMetrics{},
		&domain.DriverMetrics{},
		&domain.MerchantMetrics{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize mock repositories and services for now
	// In a real implementation, these would be proper implementations
	analyticsService := app.NewAnalyticsService(
		NewMockPlatformRepo(),
		NewMockRevenueRepo(),
		NewMockDriverRepo(),
		NewMockMerchantRepo(),
		NewMockUserService(),
		NewMockOrderService(),
		NewMockPaymentService(),
		NewMockCatalogService(),
		NewMockDriverService(),
		NewMockLocationService(),
	)

	// Setup Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "analytics-service",
		})
	})

	// Initialize HTTP handler
	handler := httpHandler.NewAnalyticsHandler(analyticsService)

	// Setup routes
	v1 := router.Group("/api/v1")
	handler.SetupRoutes(v1)

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := getEnv("ANALYTICS_SERVICE_PORT", "8011")
	log.Printf("Analytics Service starting on port %s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Mock implementations for quick startup
type mockPlatformRepo struct{}

func NewMockPlatformRepo() domain.PlatformMetricsRepository              { return &mockPlatformRepo{} }
func (m *mockPlatformRepo) Create(metrics *domain.PlatformMetrics) error { return nil }
func (m *mockPlatformRepo) GetByDate(date time.Time) (*domain.PlatformMetrics, error) {
	return nil, nil
}
func (m *mockPlatformRepo) GetByDateRange(startDate, endDate time.Time) ([]domain.PlatformMetrics, error) {
	return nil, nil
}
func (m *mockPlatformRepo) GetLatest() (*domain.PlatformMetrics, error)  { return nil, nil }
func (m *mockPlatformRepo) Update(metrics *domain.PlatformMetrics) error { return nil }

type mockRevenueRepo struct{}

func NewMockRevenueRepo() domain.RevenueMetricsRepository                           { return &mockRevenueRepo{} }
func (m *mockRevenueRepo) Create(metrics *domain.RevenueMetrics) error              { return nil }
func (m *mockRevenueRepo) GetByDate(date time.Time) (*domain.RevenueMetrics, error) { return nil, nil }
func (m *mockRevenueRepo) GetByDateRange(startDate, endDate time.Time) ([]domain.RevenueMetrics, error) {
	return nil, nil
}
func (m *mockRevenueRepo) GetTotalRevenue() (float64, error) { return 125000.50, nil }
func (m *mockRevenueRepo) GetRevenueByPeriod(startDate, endDate time.Time, interval string) ([]domain.RevenueStat, error) {
	return nil, nil
}

type mockDriverRepo struct{}

func NewMockDriverRepo() domain.DriverMetricsRepository              { return &mockDriverRepo{} }
func (m *mockDriverRepo) Create(metrics *domain.DriverMetrics) error { return nil }
func (m *mockDriverRepo) GetByDriverID(driverID string, startDate, endDate time.Time) ([]domain.DriverMetrics, error) {
	return nil, nil
}
func (m *mockDriverRepo) GetTopDrivers(period string, limit int) ([]domain.DriverSummary, error) {
	return nil, nil
}
func (m *mockDriverRepo) GetDriverPerformance(driverID string) (*domain.DriverMetrics, error) {
	return nil, nil
}
func (m *mockDriverRepo) AggregateDriverMetrics(startDate, endDate time.Time) (*domain.DriverMetrics, error) {
	return nil, nil
}

type mockMerchantRepo struct{}

func NewMockMerchantRepo() domain.MerchantMetricsRepository              { return &mockMerchantRepo{} }
func (m *mockMerchantRepo) Create(metrics *domain.MerchantMetrics) error { return nil }
func (m *mockMerchantRepo) GetByMerchantID(merchantID string, startDate, endDate time.Time) ([]domain.MerchantMetrics, error) {
	return nil, nil
}
func (m *mockMerchantRepo) GetTopMerchants(period string, limit int) ([]domain.MerchantSummary, error) {
	return nil, nil
}
func (m *mockMerchantRepo) GetMerchantPerformance(merchantID string) (*domain.MerchantMetrics, error) {
	return nil, nil
}
func (m *mockMerchantRepo) AggregateMerchantMetrics(startDate, endDate time.Time) (*domain.MerchantMetrics, error) {
	return nil, nil
}

// Mock services
type mockUserService struct{}

func NewMockUserService() domain.UserService                                { return &mockUserService{} }
func (m *mockUserService) GetUserCount() (int, error)                       { return 15420, nil }
func (m *mockUserService) GetMerchantCount() (int, error)                   { return 850, nil }
func (m *mockUserService) GetDriverCount() (int, error)                     { return 1250, nil }
func (m *mockUserService) GetUserGrowthRate(period string) (float64, error) { return 12.5, nil }

type mockOrderService struct{}

func NewMockOrderService() domain.OrderService                    { return &mockOrderService{} }
func (m *mockOrderService) GetOrderCount() (int, error)           { return 45230, nil }
func (m *mockOrderService) GetActiveOrdersCount() (int, error)    { return 89, nil }
func (m *mockOrderService) GetCompletedOrdersCount() (int, error) { return 42100, nil }
func (m *mockOrderService) GetCancelledOrdersCount() (int, error) { return 3130, nil }
func (m *mockOrderService) GetOrdersByStatus() (map[string]int, error) {
	return map[string]int{"completed": 42100, "cancelled": 3130, "active": 89}, nil
}
func (m *mockOrderService) GetOrderGrowthRate(period string) (float64, error) { return 8.3, nil }
func (m *mockOrderService) GetOrderTrends(startDate, endDate time.Time) ([]interface{}, error) {
	return nil, nil
}

type mockPaymentService struct{}

func NewMockPaymentService() domain.PaymentService                   { return &mockPaymentService{} }
func (m *mockPaymentService) GetTotalRevenue() (float64, error)      { return 1250000.75, nil }
func (m *mockPaymentService) GetDailyRevenue() (float64, error)      { return 35420.50, nil }
func (m *mockPaymentService) GetMonthlyRevenue() (float64, error)    { return 890500.25, nil }
func (m *mockPaymentService) GetAverageOrderValue() (float64, error) { return 27.65, nil }
func (m *mockPaymentService) GetRevenueByPeriod(startDate, endDate time.Time) ([]domain.RevenueStat, error) {
	return nil, nil
}
func (m *mockPaymentService) GetRevenueGrowthRate(period string) (float64, error) { return 15.2, nil }

type mockCatalogService struct{}

func NewMockCatalogService() domain.CatalogService { return &mockCatalogService{} }
func (m *mockCatalogService) GetTopMerchants(limit int) ([]domain.MerchantSummary, error) {
	return nil, nil
}
func (m *mockCatalogService) GetMerchantPerformance(merchantID string) (*domain.MerchantMetrics, error) {
	return nil, nil
}

type mockDriverService struct{}

func NewMockDriverService() domain.DriverService                                     { return &mockDriverService{} }
func (m *mockDriverService) GetTopDrivers(limit int) ([]domain.DriverSummary, error) { return nil, nil }
func (m *mockDriverService) GetDriverPerformance(driverID string) (*domain.DriverMetrics, error) {
	return nil, nil
}

type mockLocationService struct{}

func NewMockLocationService() domain.LocationService { return &mockLocationService{} }
func (m *mockLocationService) GetDriverMetrics(driverID string, startDate, endDate time.Time) (map[string]interface{}, error) {
	return nil, nil
}

package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"glovo-backend/services/location-service/internal/adapters/db"
	httpHandler "glovo-backend/services/location-service/internal/adapters/http"
	"glovo-backend/services/location-service/internal/app"
	"glovo-backend/services/location-service/internal/domain"
	"glovo-backend/shared/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// @title Location Service API
// @version 1.0
// @description Location Service for Glovo Backend
// @host localhost:8007
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Database connections
	mongoClient := database.ConnectMongoDB()

	// Initialize repositories (using the existing one)
	locationRepo := db.NewDriverLocationRepository(mongoClient)

	// For now, use mock implementations for missing repositories and services
	historyRepo := &mockLocationHistoryRepository{}
	routeRepo := &mockDeliveryRouteRepository{}
	geofenceRepo := &mockGeofenceRepository{}
	mapsService := &mockMapsService{}
	notificationService := &mockNotificationService{}

	// Initialize location service
	locationService := app.NewLocationService(
		locationRepo,
		historyRepo,
		routeRepo,
		geofenceRepo,
		mapsService,
		notificationService,
	)

	// Initialize HTTP handler
	locationHandler := httpHandler.NewLocationHandler(locationService)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Simple CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "*")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "location-service",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Location routes
		locationHandler.SetupRoutes(v1)
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := getEnv("LOCATION_SERVICE_PORT", "8008")
	log.Printf("Location Service starting on port %s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Mock implementations for missing repositories and services
type mockLocationHistoryRepository struct{}

func (m *mockLocationHistoryRepository) Create(history *domain.LocationHistory) error {
	return nil
}

func (m *mockLocationHistoryRepository) GetByDriverID(driverID string, startTime, endTime time.Time) ([]domain.LocationHistory, error) {
	return []domain.LocationHistory{}, nil
}

func (m *mockLocationHistoryRepository) GetByOrderID(orderID string) (*domain.LocationHistory, error) {
	return &domain.LocationHistory{}, nil
}

func (m *mockLocationHistoryRepository) Update(history *domain.LocationHistory) error {
	return nil
}

func (m *mockLocationHistoryRepository) Delete(id primitive.ObjectID) error {
	return nil
}

type mockDeliveryRouteRepository struct{}

func (m *mockDeliveryRouteRepository) Create(route *domain.DeliveryRoute) error {
	return nil
}

func (m *mockDeliveryRouteRepository) GetByID(id primitive.ObjectID) (*domain.DeliveryRoute, error) {
	return &domain.DeliveryRoute{}, nil
}

func (m *mockDeliveryRouteRepository) GetByOrderID(orderID string) (*domain.DeliveryRoute, error) {
	return &domain.DeliveryRoute{}, nil
}

func (m *mockDeliveryRouteRepository) GetByDriverID(driverID string) (*domain.DeliveryRoute, error) {
	return &domain.DeliveryRoute{}, nil
}

func (m *mockDeliveryRouteRepository) GetActiveRoutes() ([]domain.DeliveryRoute, error) {
	return []domain.DeliveryRoute{}, nil
}

func (m *mockDeliveryRouteRepository) Update(route *domain.DeliveryRoute) error {
	return nil
}

func (m *mockDeliveryRouteRepository) Complete(orderID string) error {
	return nil
}

func (m *mockDeliveryRouteRepository) Delete(id primitive.ObjectID) error {
	return nil
}

type mockGeofenceRepository struct{}

func (m *mockGeofenceRepository) Create(geofence *domain.Geofence) error {
	return nil
}

func (m *mockGeofenceRepository) GetByID(id primitive.ObjectID) (*domain.Geofence, error) {
	return &domain.Geofence{}, nil
}

func (m *mockGeofenceRepository) GetAll() ([]domain.Geofence, error) {
	return []domain.Geofence{}, nil
}

func (m *mockGeofenceRepository) GetByType(geofenceType domain.GeofenceType) ([]domain.Geofence, error) {
	return []domain.Geofence{}, nil
}

func (m *mockGeofenceRepository) Update(geofence *domain.Geofence) error {
	return nil
}

func (m *mockGeofenceRepository) Delete(id primitive.ObjectID) error {
	return nil
}

type mockMapsService struct{}

func (m *mockMapsService) CalculateRoute(origin, destination domain.GeoPoint) (*domain.RouteInfo, error) {
	return &domain.RouteInfo{
		Distance: 5.2,
		Duration: 15,
	}, nil
}

func (m *mockMapsService) GetETA(origin, destination domain.GeoPoint) (*domain.ETAInfo, error) {
	return &domain.ETAInfo{
		Duration:    15,
		Distance:    5.2,
		ArrivalTime: time.Now().Add(15 * time.Minute),
	}, nil
}

func (m *mockMapsService) ReverseGeocode(location domain.GeoPoint) (*domain.AddressInfo, error) {
	return &domain.AddressInfo{
		Address: "Mock Address",
		City:    "Mock City",
	}, nil
}

type mockNotificationService struct{}

func (m *mockNotificationService) SendGeofenceAlert(event domain.GeofenceEvent) error {
	log.Printf("Mock geofence alert sent for driver %s", event.DriverID)
	return nil
}

func (m *mockNotificationService) SendLocationAlert(driverID string, message string) error {
	log.Printf("Mock location alert sent to driver %s: %s", driverID, message)
	return nil
}

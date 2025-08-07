package main

import (
	"log"
	"net/http"
	"os"

	"glovo-backend/services/delivery-service/internal/adapters/client"
	"glovo-backend/services/delivery-service/internal/adapters/db"
	httpHandler "glovo-backend/services/delivery-service/internal/adapters/http"
	"glovo-backend/services/delivery-service/internal/app"
	"glovo-backend/services/delivery-service/internal/domain"
	"glovo-backend/shared/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Delivery Service API
// @version 1.0
// @description Delivery assignment and management service
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8004
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize database connection
	postgresDB := database.ConnectPostgres()

	// Auto-migrate database schema
	if err := postgresDB.AutoMigrate(
		&domain.Delivery{},
		&domain.DeliveryAssignment{},
		&domain.DriverPerformance{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize repositories
	deliveryRepo := db.NewDeliveryRepository(postgresDB)
	assignmentRepo := db.NewDeliveryAssignmentRepository(postgresDB)
	performanceRepo := db.NewDriverPerformanceRepository(postgresDB)

	// Initialize external service clients (mock for now)
	orderService := client.NewMockOrderService()
	driverService := client.NewMockDriverService()
	locationService := client.NewMockLocationService()
	notificationService := client.NewMockNotificationService()
	paymentService := client.NewMockPaymentService()

	// Initialize use case
	deliveryService := app.NewDeliveryService(
		deliveryRepo,
		assignmentRepo,
		performanceRepo,
		orderService,
		driverService,
		locationService,
		notificationService,
		paymentService,
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
			"service": "delivery-service",
		})
	})

	// Initialize HTTP handler
	handler := httpHandler.NewDeliveryHandler(deliveryService)

	// Setup routes
	v1 := router.Group("/api/v1")
	handler.SetupRoutes(v1)

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := getEnv("PORT", "8004")
	log.Printf("Delivery Service starting on port %s", port)

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

package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"glovo-backend/services/admin-service/internal/adapters/client"
	"glovo-backend/services/admin-service/internal/adapters/db"
	httpHandler "glovo-backend/services/admin-service/internal/adapters/http"
	"glovo-backend/services/admin-service/internal/app"
	"glovo-backend/services/admin-service/internal/domain"
	"glovo-backend/shared/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Admin Service API
// @version 1.0
// @description Admin Service for Glovo Backend
// @host localhost:8009
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
	postgresDB := database.ConnectPostgres()

	// Auto-migrate database schemas
	if err := postgresDB.AutoMigrate(
		&domain.Admin{},
		&domain.SystemConfig{},
		&domain.AuditLog{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize repositories
	adminRepo := db.NewAdminRepository(postgresDB)
	systemConfigRepo := db.NewSystemConfigRepository(postgresDB)
	auditLogRepo := db.NewAuditLogRepository(postgresDB)

	// Initialize external service clients (using mocks for development)
	userService := client.NewMockUserService()
	orderService := client.NewMockOrderService()
	paymentService := client.NewMockPaymentService()
	catalogService := client.NewMockCatalogService()
	driverService := client.NewMockDriverService()
	analyticsService := client.NewMockAnalyticsService()

	// Initialize admin service
	adminService := app.NewAdminService(
		adminRepo,
		systemConfigRepo,
		auditLogRepo,
		userService,
		orderService,
		paymentService,
		catalogService,
		driverService,
		analyticsService,
	)

	// Initialize HTTP handler
	adminHandler := httpHandler.NewAdminHandler(adminService)

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
			"service": "admin-service",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Admin routes
		adminHandler.SetupRoutes(v1)
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := getEnv("ADMIN_SERVICE_PORT", "8010")
	log.Printf("Admin Service starting on port %s", port)

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

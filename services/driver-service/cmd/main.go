package main

import (
	"log"
	"net/http"
	"os"

	"glovo-backend/services/driver-service/internal/adapters/client"
	"glovo-backend/services/driver-service/internal/adapters/db"
	httpHandler "glovo-backend/services/driver-service/internal/adapters/http"
	"glovo-backend/services/driver-service/internal/app"
	"glovo-backend/services/driver-service/internal/domain"
	"glovo-backend/shared/database"
	"glovo-backend/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Driver Service API
// @version 1.0
// @description Driver registration, management and performance tracking service
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8005
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
		&domain.Driver{},
		&domain.DriverDocument{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize repositories
	driverRepo := db.NewDriverRepository(postgresDB)
	documentRepo := db.NewDriverDocumentRepository(postgresDB)

	// Initialize external service clients (mock for now)
	userService := client.NewMockUserService()
	locationService := client.NewMockLocationService()
	paymentService := client.NewMockPaymentService()

	// Initialize use case
	driverService := app.NewDriverService(
		driverRepo,
		documentRepo,
		userService,
		locationService,
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
			"service": "driver-service",
		})
	})

	// Initialize HTTP handler
	handler := httpHandler.NewDriverHandler(driverService)

	// Setup routes
	v1 := router.Group("/api/v1")
	handler.SetupRoutes(v1)

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// TODO: Remove the inline route handlers below (kept for reference)
	// v1_old := router.Group("/api/v1/old")
	{
		// Public driver registration (after user authentication)
		v1.POST("/drivers/register", middleware.AuthMiddleware(), func(c *gin.Context) {
			userID := c.GetString("user_id")

			var req domain.RegisterDriverRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			driver, err := driverService.RegisterDriver(userID, req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, driver)
		})

		// Driver profile routes (authenticated drivers)
		drivers := v1.Group("/drivers")
		drivers.Use(middleware.AuthMiddleware())
		{
			drivers.GET("/profile", func(c *gin.Context) {
				userID := c.GetString("user_id")

				driver, err := driverService.GetDriverByUser(userID)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Driver profile not found"})
					return
				}

				c.JSON(http.StatusOK, driver)
			})

			drivers.PUT("/profile", func(c *gin.Context) {
				userID := c.GetString("user_id")

				// Get driver ID from user
				driver, err := driverService.GetDriverByUser(userID)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Driver profile not found"})
					return
				}

				var req domain.UpdateDriverProfileRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				updatedDriver, err := driverService.UpdateProfile(driver.ID, userID, req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, updatedDriver)
			})

			drivers.PUT("/status", func(c *gin.Context) {
				userID := c.GetString("user_id")

				driver, err := driverService.GetDriverByUser(userID)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Driver profile not found"})
					return
				}

				var req domain.UpdateStatusRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				updatedDriver, err := driverService.UpdateStatus(driver.ID, userID, req.Status)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, updatedDriver)
			})

			drivers.PUT("/location", func(c *gin.Context) {
				userID := c.GetString("user_id")

				driver, err := driverService.GetDriverByUser(userID)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Driver profile not found"})
					return
				}

				var req domain.UpdateLocationRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				updatedDriver, err := driverService.UpdateLocation(driver.ID, userID, req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, updatedDriver)
			})

			// Document management
			drivers.POST("/documents", func(c *gin.Context) {
				userID := c.GetString("user_id")

				driver, err := driverService.GetDriverByUser(userID)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Driver profile not found"})
					return
				}

				var req domain.UploadDocumentRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				document, err := driverService.UploadDocument(driver.ID, userID, req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, document)
			})

			drivers.GET("/documents", func(c *gin.Context) {
				userID := c.GetString("user_id")

				driver, err := driverService.GetDriverByUser(userID)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Driver profile not found"})
					return
				}

				documents, err := driverService.GetDocuments(driver.ID, userID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, documents)
			})
		}

		// Public search endpoints
		v1.GET("/drivers/search", func(c *gin.Context) {
			var req domain.DriverSearchRequest
			if err := c.ShouldBindQuery(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			drivers, err := driverService.SearchDrivers(req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, drivers)
		})
	}

	// Start server
	port := getEnv("PORT", "8005")
	log.Printf("Driver Service starting on port %s", port)

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

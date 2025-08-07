package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"glovo-backend/services/notification-service/internal/adapters/client"
	"glovo-backend/services/notification-service/internal/adapters/db"
	httpHandler "glovo-backend/services/notification-service/internal/adapters/http"
	"glovo-backend/services/notification-service/internal/app"
	"glovo-backend/services/notification-service/internal/domain"
	"glovo-backend/shared/auth"
	"glovo-backend/shared/database"
	"glovo-backend/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Notification Service API
// @version 1.0
// @description SMS, Email, Push notifications and template management service
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8008
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
		&domain.Notification{},
		&domain.NotificationTemplate{},
		&domain.UserPreference{},
		&domain.NotificationDevice{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize repositories
	notificationRepo := db.NewNotificationRepository(postgresDB)
	templateRepo := db.NewTemplateRepository(postgresDB)
	preferenceRepo := db.NewPreferenceRepository(postgresDB)
	deviceRepo := db.NewDeviceRepository(postgresDB)

	// Initialize external service clients (mock for now)
	pushService := client.NewMockPushNotificationService()
	smsService := client.NewMockSMSService()
	emailService := client.NewMockEmailService()

	// Initialize use case
	notificationService := app.NewNotificationService(
		notificationRepo,
		templateRepo,
		preferenceRepo,
		deviceRepo,
		pushService,
		smsService,
		emailService,
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
			"service": "notification-service",
		})
	})

	// Initialize HTTP handler
	handler := httpHandler.NewNotificationHandler(notificationService)

	// Setup routes
	v1 := router.Group("/api/v1")
	handler.SetupRoutes(v1)

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// TODO: Remove the inline route handlers below (kept for reference)
	// v1_old := router.Group("/api/v1/old")
	{
		// Public OTP endpoint (for User Service)
		v1.POST("/notifications/otp", func(c *gin.Context) {
			var req struct {
				PhoneNumber string `json:"phone_number" binding:"required"`
				OTPCode     string `json:"otp_code" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err := notificationService.SendOTPNotification(req.PhoneNumber, req.OTPCode)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
		})

		// Order notification endpoint (for Order Service)
		v1.POST("/notifications/order", func(c *gin.Context) {
			var req struct {
				OrderID string `json:"order_id" binding:"required"`
				UserID  string `json:"user_id" binding:"required"`
				Message string `json:"message" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			err := notificationService.SendOrderNotification(req.OrderID, req.UserID, req.Message)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Order notification sent"})
		})

		// Authenticated notification routes
		notifications := v1.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware())
		{
			// Send single notification
			notifications.POST("/send", func(c *gin.Context) {
				var req domain.SendNotificationRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				notification, err := notificationService.SendNotification(req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, notification)
			})

			// Send bulk notifications (admin only)
			notifications.POST("/bulk", middleware.RequireRole(auth.RoleAdmin), func(c *gin.Context) {
				var req domain.BulkNotificationRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				notifications, err := notificationService.SendBulkNotification(req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, notifications)
			})

			// Send template notification
			notifications.POST("/template", func(c *gin.Context) {
				var req domain.SendTemplateNotificationRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				notification, err := notificationService.SendTemplateNotification(req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, notification)
			})

			// Get user notifications
			notifications.GET("/", func(c *gin.Context) {
				userID := c.GetString("user_id")

				limitStr := c.DefaultQuery("limit", "20")
				offsetStr := c.DefaultQuery("offset", "0")

				limit, _ := strconv.Atoi(limitStr)
				offset, _ := strconv.Atoi(offsetStr)

				response, err := notificationService.GetNotifications(userID, limit, offset)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, response)
			})

			// Get specific notification
			notifications.GET("/:id", func(c *gin.Context) {
				notificationID := c.Param("id")

				notification, err := notificationService.GetNotification(notificationID)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
					return
				}

				c.JSON(http.StatusOK, notification)
			})

			// Mark notification as read
			notifications.PUT("/:id/read", func(c *gin.Context) {
				userID := c.GetString("user_id")
				notificationID := c.Param("id")

				err := notificationService.MarkNotificationAsRead(notificationID, userID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
			})

			// Mark all notifications as read
			notifications.PUT("/read-all", func(c *gin.Context) {
				userID := c.GetString("user_id")

				err := notificationService.MarkAllNotificationsAsRead(userID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
			})

			// Get unread count
			notifications.GET("/unread/count", func(c *gin.Context) {
				userID := c.GetString("user_id")

				count, err := notificationService.GetUnreadCount(userID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"unread_count": count})
			})
		}

		// Template management (admin only)
		templates := v1.Group("/templates")
		templates.Use(middleware.AuthMiddleware())
		templates.Use(middleware.RequireRole(auth.RoleAdmin))
		{
			templates.POST("/", func(c *gin.Context) {
				var template domain.NotificationTemplate
				if err := c.ShouldBindJSON(&template); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				createdTemplate, err := notificationService.CreateTemplate(&template)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, createdTemplate)
			})

			templates.GET("/", func(c *gin.Context) {
				limitStr := c.DefaultQuery("limit", "20")
				offsetStr := c.DefaultQuery("offset", "0")

				limit, _ := strconv.Atoi(limitStr)
				offset, _ := strconv.Atoi(offsetStr)

				templates, err := notificationService.ListTemplates(limit, offset)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, templates)
			})

			templates.GET("/:id", func(c *gin.Context) {
				templateID := c.Param("id")

				template, err := notificationService.GetTemplate(templateID)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
					return
				}

				c.JSON(http.StatusOK, template)
			})

			templates.PUT("/:id", func(c *gin.Context) {
				templateID := c.Param("id")

				var updates map[string]interface{}
				if err := c.ShouldBindJSON(&updates); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				updatedTemplate, err := notificationService.UpdateTemplate(templateID, updates)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, updatedTemplate)
			})

			templates.DELETE("/:id", func(c *gin.Context) {
				templateID := c.Param("id")

				err := notificationService.DeleteTemplate(templateID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "Template deleted"})
			})
		}

		// User preferences
		preferences := v1.Group("/preferences")
		preferences.Use(middleware.AuthMiddleware())
		{
			preferences.GET("/", func(c *gin.Context) {
				userID := c.GetString("user_id")

				prefs, err := notificationService.GetUserPreferences(userID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, prefs)
			})

			preferences.PUT("/", func(c *gin.Context) {
				userID := c.GetString("user_id")

				var req domain.UpdatePreferenceRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				preference, err := notificationService.UpdateUserPreference(userID, req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, preference)
			})
		}

		// Device management
		devices := v1.Group("/devices")
		devices.Use(middleware.AuthMiddleware())
		{
			devices.POST("/register", func(c *gin.Context) {
				userID := c.GetString("user_id")

				var req domain.RegisterDeviceRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				device, err := notificationService.RegisterDevice(userID, req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, device)
			})

			devices.GET("/", func(c *gin.Context) {
				userID := c.GetString("user_id")

				devices, err := notificationService.GetUserDevices(userID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, devices)
			})

			devices.DELETE("/:token", func(c *gin.Context) {
				userID := c.GetString("user_id")
				deviceToken := c.Param("token")

				err := notificationService.DeactivateDevice(userID, deviceToken)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "Device deactivated"})
			})
		}
	}

	// Start server
	port := getEnv("NOTIFICATION_SERVICE_PORT", "8009")
	log.Printf("Notification Service starting on port %s", port)

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

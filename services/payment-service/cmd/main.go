package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"glovo-backend/services/payment-service/internal/adapters/client"
	"glovo-backend/services/payment-service/internal/adapters/db"
	httpHandler "glovo-backend/services/payment-service/internal/adapters/http"
	"glovo-backend/services/payment-service/internal/app"
	"glovo-backend/services/payment-service/internal/domain"
	"glovo-backend/shared/auth"
	"glovo-backend/shared/database"
	"glovo-backend/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Payment Service API
// @version 1.0
// @description Wallet management, payment processing and commission service
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8006
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
		&domain.Wallet{},
		&domain.Transaction{},
		&domain.PaymentMethod{},
		&domain.Commission{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize repositories
	walletRepo := db.NewWalletRepository(postgresDB)
	transactionRepo := db.NewTransactionRepository(postgresDB)
	paymentMethodRepo := db.NewPaymentMethodRepository(postgresDB)
	commissionRepo := db.NewCommissionRepository(postgresDB)

	// Initialize external service clients (mock for now)
	stripeService := client.NewMockStripeService()
	bankService := client.NewMockBankService()

	// Initialize use case
	paymentService := app.NewPaymentService(
		walletRepo,
		transactionRepo,
		paymentMethodRepo,
		commissionRepo,
		stripeService,
		bankService,
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
			"service": "payment-service",
		})
	})

	// Initialize HTTP handler
	handler := httpHandler.NewPaymentHandler(paymentService)

	// Setup routes
	v1 := router.Group("/api/v1")
	handler.SetupRoutes(v1)

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// TODO: Remove the inline route handlers below (kept for reference)
	// v1_old := router.Group("/api/v1/old")
	{
		// Wallet routes (authenticated)
		wallets := v1.Group("/wallets")
		wallets.Use(middleware.AuthMiddleware())
		{
			wallets.POST("/create", func(c *gin.Context) {
				userID := c.GetString("user_id")
				role := c.GetString("role")

				var userRole auth.UserRole
				switch role {
				case "customer":
					userRole = auth.RoleCustomer
				case "merchant":
					userRole = auth.RoleMerchant
				case "driver":
					userRole = auth.RoleDriver
				default:
					userRole = auth.RoleCustomer
				}

				wallet, err := paymentService.CreateWallet(userID, userRole)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, wallet)
			})

			wallets.GET("/", func(c *gin.Context) {
				userID := c.GetString("user_id")

				wallet, err := paymentService.GetWallet(userID)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
					return
				}

				c.JSON(http.StatusOK, wallet)
			})

			wallets.GET("/balance", func(c *gin.Context) {
				userID := c.GetString("user_id")

				balance, err := paymentService.GetBalance(userID)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
					return
				}

				c.JSON(http.StatusOK, balance)
			})
		}

		// Payment processing routes (authenticated)
		payments := v1.Group("/payments")
		payments.Use(middleware.AuthMiddleware())
		{
			payments.POST("/process", func(c *gin.Context) {
				var req domain.ProcessPaymentRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				response, err := paymentService.ProcessPayment(req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, response)
			})

			payments.POST("/refund", func(c *gin.Context) {
				var req domain.RefundRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				response, err := paymentService.ProcessRefund(req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, response)
			})

			payments.POST("/transfer", func(c *gin.Context) {
				var req domain.TransferRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				response, err := paymentService.ProcessTransfer(req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, response)
			})

			payments.POST("/topup", func(c *gin.Context) {
				var req domain.TopUpRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				response, err := paymentService.ProcessTopUp(req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, response)
			})

			payments.POST("/withdraw", func(c *gin.Context) {
				var req domain.WithdrawalRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				response, err := paymentService.ProcessWithdrawal(req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, response)
			})
		}

		// Payment methods routes (authenticated)
		methods := v1.Group("/payment-methods")
		methods.Use(middleware.AuthMiddleware())
		{
			methods.POST("/", func(c *gin.Context) {
				userID := c.GetString("user_id")

				var req domain.AddPaymentMethodRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				method, err := paymentService.AddPaymentMethod(userID, req)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, method)
			})

			methods.GET("/", func(c *gin.Context) {
				userID := c.GetString("user_id")

				methods, err := paymentService.GetPaymentMethods(userID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, methods)
			})

			methods.PUT("/:id/default", func(c *gin.Context) {
				userID := c.GetString("user_id")
				methodID := c.Param("id")

				err := paymentService.SetDefaultPaymentMethod(userID, methodID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "Payment method set as default"})
			})

			methods.DELETE("/:id", func(c *gin.Context) {
				userID := c.GetString("user_id")
				methodID := c.Param("id")

				err := paymentService.RemovePaymentMethod(userID, methodID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{"message": "Payment method removed"})
			})
		}

		// Transaction routes (authenticated)
		transactions := v1.Group("/transactions")
		transactions.Use(middleware.AuthMiddleware())
		{
			transactions.GET("/:id", func(c *gin.Context) {
				transactionID := c.Param("id")

				transaction, err := paymentService.GetTransaction(transactionID)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
					return
				}

				c.JSON(http.StatusOK, transaction)
			})

			transactions.GET("/", func(c *gin.Context) {
				userID := c.GetString("user_id")

				limitStr := c.DefaultQuery("limit", "20")
				offsetStr := c.DefaultQuery("offset", "0")

				limit, _ := strconv.Atoi(limitStr)
				offset, _ := strconv.Atoi(offsetStr)

				transactions, err := paymentService.GetTransactionHistory(userID, limit, offset)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, transactions)
			})

			transactions.GET("/report", func(c *gin.Context) {
				userID := c.GetString("user_id")

				startDateStr := c.Query("start_date")
				endDateStr := c.Query("end_date")

				startDate, err := time.Parse("2006-01-02", startDateStr)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (YYYY-MM-DD)"})
					return
				}

				endDate, err := time.Parse("2006-01-02", endDateStr)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (YYYY-MM-DD)"})
					return
				}

				report, err := paymentService.GetTransactionReport(userID, startDate, endDate)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, report)
			})
		}

		// Commission routes (admin/merchant access)
		commissions := v1.Group("/commissions")
		commissions.Use(middleware.AuthMiddleware())
		{
			commissions.POST("/calculate", middleware.RequireRoles([]auth.UserRole{auth.RoleAdmin}), func(c *gin.Context) {
				var req struct {
					OrderID     string  `json:"order_id" binding:"required"`
					OrderAmount float64 `json:"order_amount" binding:"required"`
					DeliveryFee float64 `json:"delivery_fee" binding:"required"`
					MerchantID  string  `json:"merchant_id" binding:"required"`
					DriverID    string  `json:"driver_id" binding:"required"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				commission, err := paymentService.CalculateCommission(
					req.OrderID, req.OrderAmount, req.DeliveryFee, req.MerchantID, req.DriverID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusCreated, commission)
			})

			commissions.GET("/merchant", middleware.RequireRoles([]auth.UserRole{auth.RoleMerchant, auth.RoleAdmin}), func(c *gin.Context) {
				userID := c.GetString("user_id")
				role := c.GetString("role")

				var merchantID string
				if role == "admin" {
					merchantID = c.Query("merchant_id")
					if merchantID == "" {
						c.JSON(http.StatusBadRequest, gin.H{"error": "merchant_id required for admin"})
						return
					}
				} else {
					merchantID = userID
				}

				limitStr := c.DefaultQuery("limit", "20")
				offsetStr := c.DefaultQuery("offset", "0")

				limit, _ := strconv.Atoi(limitStr)
				offset, _ := strconv.Atoi(offsetStr)

				commissions, err := paymentService.GetMerchantCommissions(merchantID, limit, offset)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, commissions)
			})

			commissions.GET("/driver", middleware.RequireRoles([]auth.UserRole{auth.RoleDriver, auth.RoleAdmin}), func(c *gin.Context) {
				userID := c.GetString("user_id")
				role := c.GetString("role")

				var driverID string
				if role == "admin" {
					driverID = c.Query("driver_id")
					if driverID == "" {
						c.JSON(http.StatusBadRequest, gin.H{"error": "driver_id required for admin"})
						return
					}
				} else {
					driverID = userID
				}

				limitStr := c.DefaultQuery("limit", "20")
				offsetStr := c.DefaultQuery("offset", "0")

				limit, _ := strconv.Atoi(limitStr)
				offset, _ := strconv.Atoi(offsetStr)

				commissions, err := paymentService.GetDriverCommissions(driverID, limit, offset)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, commissions)
			})
		}

		// Payout routes (admin only)
		payouts := v1.Group("/payouts")
		payouts.Use(middleware.AuthMiddleware())
		payouts.Use(middleware.RequireRole(auth.RoleAdmin))
		{
			payouts.POST("/merchant", func(c *gin.Context) {
				var req struct {
					MerchantID string  `json:"merchant_id" binding:"required"`
					Amount     float64 `json:"amount" binding:"required"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				response, err := paymentService.ProcessMerchantPayout(req.MerchantID, req.Amount)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, response)
			})

			payouts.POST("/driver", func(c *gin.Context) {
				var req struct {
					DriverID string  `json:"driver_id" binding:"required"`
					Amount   float64 `json:"amount" binding:"required"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				response, err := paymentService.ProcessDriverPayout(req.DriverID, req.Amount)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, response)
			})
		}
	}

	// Start server
	port := getEnv("PAYMENT_SERVICE_PORT", "8007")
	log.Printf("Payment Service starting on port %s", port)

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

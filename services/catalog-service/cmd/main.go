package main

import (
	"log"
	"net/http"
	"os"

	"glovo-backend/services/catalog-service/internal/adapters/db"
	httpAdapter "glovo-backend/services/catalog-service/internal/adapters/http"
	"glovo-backend/services/catalog-service/internal/app"
	"glovo-backend/services/catalog-service/internal/domain"
	"glovo-backend/shared/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Catalog Service API
// @version 1.0
// @description Store and product catalog management service
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8003
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
		&domain.Store{},
		&domain.Product{},
		&domain.ProductOption{},
		&domain.ProductOptionChoice{},
		&domain.Category{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize repositories
	storeRepo := db.NewStoreRepository(postgresDB)
	productRepo := db.NewProductRepository(postgresDB)
	categoryRepo := db.NewCategoryRepository(postgresDB)

	// Initialize use case
	catalogService := app.NewCatalogService(storeRepo, productRepo, categoryRepo)

	// Initialize HTTP handler
	catalogHandler := httpAdapter.NewCatalogHandler(catalogService)

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
			"service": "catalog-service",
		})
	})

	// Setup routes
	catalogHandler.SetupRoutes(router)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := getEnv("PORT", "8003")
	log.Printf("Catalog Service starting on port %s", port)
	log.Printf("Swagger documentation available at http://localhost:%s/swagger/index.html", port)

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

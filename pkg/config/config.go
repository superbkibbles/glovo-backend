package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// PostgreSQL
	DatabaseURL string

	// MongoDB (for notification, file services)
	MongoURI string
	MongoDB  string

	// Redis
	RedisURI      string
	RedisPassword string

	// JWT
	JWTSecret        string
	JWTExpiry        time.Duration
	JWTRefreshExpiry time.Duration

	// Password
	PasswordPepper string

	// Superadmin
	SuperadminUsername string
	SuperadminPassword string

	// Service Ports (gRPC)
	AuthServicePort         string
	UserServicePort         string
	NotificationServicePort string
	FileStorageServicePort  string
	RestaurantServicePort   string
	OrderServicePort        string
	DeliveryServicePort     string
	SettingsServicePort     string

	// Gateway
	GatewayHTTPPort string
	GatewayBaseURL  string

	// Service Addresses
	AuthServiceAddr         string
	UserServiceAddr         string
	NotificationServiceAddr string
	FileStorageServiceAddr  string
	RestaurantServiceAddr   string
	OrderServiceAddr        string
	DeliveryServiceAddr     string
	SettingsServiceAddr     string

	// File Storage
	UploadPath    string
	MaxUploadSize int64

	// Firebase
	FirebaseProjectID       string
	FirebaseCredentialsPath string

	// Environment
	Env string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		// PostgreSQL
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/food_delivery?sslmode=disable"),

		// MongoDB
		MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:  getEnv("MONGO_DB", "food_delivery"),

		// Redis
		RedisURI:      getEnv("REDIS_URI", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		// JWT
		JWTSecret:        getEnv("JWT_SECRET", "default-secret-change-me"),
		JWTExpiry:        getDurationEnv("JWT_EXPIRY", 24*time.Hour),
		JWTRefreshExpiry: getDurationEnv("JWT_REFRESH_EXPIRY", 7*24*time.Hour),

		// Password
		PasswordPepper: getEnv("PASSWORD_PEPPER", "default-pepper-change-me"),

		// Superadmin
		SuperadminUsername: getEnv("SUPERADMIN_USERNAME", "admin"),
		SuperadminPassword:  getEnv("SUPERADMIN_PASSWORD", "Admin@123"),

		// Service Ports (gRPC)
		AuthServicePort:         getEnv("AUTH_SERVICE_GRPC_PORT", "50051"),
		UserServicePort:           getEnv("USER_SERVICE_GRPC_PORT", "50052"),
		NotificationServicePort:   getEnv("NOTIFICATION_SERVICE_GRPC_PORT", "50053"),
		FileStorageServicePort:    getEnv("FILE_STORAGE_SERVICE_GRPC_PORT", "50054"),
		RestaurantServicePort:     getEnv("RESTAURANT_SERVICE_GRPC_PORT", "50055"),
		OrderServicePort:          getEnv("ORDER_SERVICE_GRPC_PORT", "50056"),
		DeliveryServicePort:       getEnv("DELIVERY_SERVICE_GRPC_PORT", "50057"),
		SettingsServicePort:       getEnv("SETTINGS_SERVICE_GRPC_PORT", "50058"),

		// Gateway
		GatewayHTTPPort: getEnv("GATEWAY_HTTP_PORT", "8080"),
		GatewayBaseURL:  getEnv("GATEWAY_BASE_URL", "http://localhost:8080"),

		// Service Addresses
		AuthServiceAddr:         getEnv("AUTH_SERVICE_ADDR", "localhost:50051"),
		UserServiceAddr:         getEnv("USER_SERVICE_ADDR", "localhost:50052"),
		NotificationServiceAddr: getEnv("NOTIFICATION_SERVICE_ADDR", "localhost:50053"),
		FileStorageServiceAddr:  getEnv("FILE_STORAGE_SERVICE_ADDR", "localhost:50054"),
		RestaurantServiceAddr:   getEnv("RESTAURANT_SERVICE_ADDR", "localhost:50055"),
		OrderServiceAddr:        getEnv("ORDER_SERVICE_ADDR", "localhost:50056"),
		DeliveryServiceAddr:     getEnv("DELIVERY_SERVICE_ADDR", "localhost:50057"),
		SettingsServiceAddr:     getEnv("SETTINGS_SERVICE_ADDR", "localhost:50058"),

		// File Storage
		UploadPath:    getEnv("UPLOAD_PATH", "./uploads"),
		MaxUploadSize: getInt64Env("MAX_UPLOAD_SIZE", 52428800),

		// Firebase
		FirebaseProjectID:       getEnv("FIREBASE_PROJECT_ID", "food-delivery-4479f"),
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", "./food-delivery-4479f-firebase-adminsdk-fbsvc-e96b9be0ce.json"),

		// Environment
		Env: getEnv("ENV", "development"),
	}
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

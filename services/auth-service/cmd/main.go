package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mendmzury/food-delivery/pkg/config"
	"github.com/mendmzury/food-delivery/pkg/database"
	"github.com/mendmzury/food-delivery/pkg/logger"
	"github.com/mendmzury/food-delivery/pkg/middleware"
	"github.com/mendmzury/food-delivery/services/auth-service/internal/application"
	"github.com/mendmzury/food-delivery/services/auth-service/internal/domain"
	"github.com/mendmzury/food-delivery/services/auth-service/internal/grpc/server"
	"github.com/mendmzury/food-delivery/pkg/utils"
	"github.com/google/uuid"
	"github.com/mendmzury/food-delivery/services/auth-service/internal/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.Env)
	log := logger.WithService("auth-service")

	log.Info().Msg("Starting auth-service...")

	// Connect to Postgres
	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Postgres")
	}

	// Auto migrate
	if err := db.AutoMigrate(&domain.User{}, &domain.Role{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}

	// Connect to Redis
	redisDB, err := database.NewRedisDB(cfg.RedisURI, cfg.RedisPassword)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer func() {
		if err := redisDB.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close Redis connection")
		}
	}()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	sessionRepo := repository.NewSessionRepository(redisDB)
	otpRepo := repository.NewOTPRepository(redisDB)

	// Ensure superadmin exists
	ensureSuperadmin(db, cfg)

	// Initialize application service
	authService := application.NewAuthService(cfg, userRepo, roleRepo, sessionRepo, otpRepo)

	// Create auth interceptor
	authInterceptor := middleware.NewAuthInterceptor(cfg.JWTSecret)
	publicMethods := []string{
		"/auth.AuthService/Login",
		"/auth.AuthService/ForgotPassword",
		"/auth.AuthService/RefreshToken",
		"/auth.AuthService/SignUpWithPhone",
		"/auth.AuthService/SignUpWithGoogle",
		"/auth.AuthService/SignUpWithEmail",
		"/auth.AuthService/SignInWithPhone",
		"/auth.AuthService/SignInWithEmail",
		"/auth.AuthService/SendOTP",
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryServerInterceptor(publicMethods)),
	)
	server.RegisterAuthServer(grpcServer, authService)

	if cfg.IsDevelopment() {
		reflection.Register(grpcServer)
	}

	address := fmt.Sprintf(":%s", cfg.AuthServicePort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Str("address", address).Msg("Failed to listen")
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info().Msg("Shutting down auth-service...")
		grpcServer.GracefulStop()
	}()

	log.Info().Str("address", address).Msg("auth-service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve")
	}
}

func ensureSuperadmin(db *gorm.DB, cfg *config.Config) {
	ctx := context.Background()
	// Check if superadmin role exists
	var roleCount int64
	db.WithContext(ctx).Model(&domain.Role{}).Count(&roleCount)
	if roleCount == 0 {
		// Create admin role
		adminRole := &domain.Role{
			ID:          uuid.New(),
			Name:        "Admin",
			Icon:        "shield",
			Permissions: domain.StringArray(middleware.AllPermissions()),
		}
		if err := db.WithContext(ctx).Create(adminRole).Error; err != nil {
			return
		}
	}
	// Check if superadmin user exists
	var user domain.User
	if err := db.WithContext(ctx).Where("is_superadmin = ?", true).First(&user).Error; err == nil {
		return // superadmin exists
	}
	// Create superadmin - need to get admin role
	var adminRole domain.Role
	if err := db.WithContext(ctx).Where("name = ?", "Admin").First(&adminRole).Error; err != nil {
		return
	}
	pwHasher := utils.NewPasswordHasher(cfg.PasswordPepper)
	salt, err := pwHasher.GenerateSalt()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to generate salt for superadmin")
		return
	}
	hash, err := pwHasher.HashPassword(cfg.SuperadminPassword, salt)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to hash password for superadmin")
		return
	}
	superadmin := &domain.User{
		ID:           uuid.New(),
		Username:     cfg.SuperadminUsername,
		PasswordHash: hash,
		Salt:         salt,
		Name:         "Superadmin",
		RoleID:       adminRole.ID,
		UserType:     "admin",
		IsSuperadmin: true,
		Active:       true,
	}
	db.WithContext(ctx).Create(superadmin)
}

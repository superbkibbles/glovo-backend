package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/mendmzury/food-delivery/pkg/config"
	"github.com/mendmzury/food-delivery/pkg/database"
	"github.com/mendmzury/food-delivery/pkg/logger"
	"github.com/mendmzury/food-delivery/pkg/middleware"
	"github.com/mendmzury/food-delivery/services/user-service/internal/application"
	"github.com/mendmzury/food-delivery/services/user-service/internal/domain"
	"github.com/mendmzury/food-delivery/services/user-service/internal/grpc/server"
	"github.com/mendmzury/food-delivery/services/user-service/internal/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.Env)
	log := logger.WithService("user-service")

	log.Info().Msg("Starting user-service...")

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Postgres")
	}

	// Auto migrate - uses same tables as auth-service
	if err := db.AutoMigrate(&domain.User{}, &domain.Role{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	userService := application.NewUserService(userRepo, roleRepo, cfg.PasswordPepper)

	authInterceptor := middleware.NewAuthInterceptor(cfg.JWTSecret)
	publicMethods := []string{} // All methods require authentication

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.UnaryServerInterceptor(publicMethods)),
	)
	server.RegisterUserServer(grpcServer, userService)

	if cfg.IsDevelopment() {
		reflection.Register(grpcServer)
	}

	address := fmt.Sprintf(":%s", cfg.UserServicePort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Str("address", address).Msg("Failed to listen")
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info().Msg("Shutting down user-service...")
		grpcServer.GracefulStop()
	}()

	log.Info().Str("address", address).Msg("user-service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve")
	}
}

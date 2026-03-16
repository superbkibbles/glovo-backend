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
	"github.com/mendmzury/food-delivery/services/restaurant-service/internal/application"
	"github.com/mendmzury/food-delivery/services/restaurant-service/internal/domain"
	"github.com/mendmzury/food-delivery/services/restaurant-service/internal/grpc/server"
	"github.com/mendmzury/food-delivery/services/restaurant-service/internal/repository"
	restaurantpb "github.com/mendmzury/food-delivery/proto/restaurant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.Env)
	log := logger.WithService("restaurant-service")

	log.Info().Msg("Starting restaurant-service...")

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Postgres")
	}

	if err := db.AutoMigrate(&domain.Restaurant{}, &domain.Category{}, &domain.MenuItem{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}

	restaurantRepo := repository.NewRestaurantRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	menuItemRepo := repository.NewMenuItemRepository(db)
	restaurantService := application.NewRestaurantService(restaurantRepo, categoryRepo, menuItemRepo)

	srv := server.NewRestaurantServer(restaurantService)
	grpcServer := grpc.NewServer()
	restaurantpb.RegisterRestaurantServiceServer(grpcServer, srv)

	if cfg.IsDevelopment() {
		reflection.Register(grpcServer)
	}

	address := fmt.Sprintf(":%s", cfg.RestaurantServicePort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Str("address", address).Msg("Failed to listen")
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info().Msg("Shutting down restaurant-service...")
		grpcServer.GracefulStop()
	}()

	log.Info().Str("address", address).Msg("restaurant-service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve")
	}
}

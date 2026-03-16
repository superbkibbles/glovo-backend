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
	"github.com/mendmzury/food-delivery/services/delivery-service/internal/application"
	"github.com/mendmzury/food-delivery/services/delivery-service/internal/domain"
	"github.com/mendmzury/food-delivery/services/delivery-service/internal/grpc/server"
	"github.com/mendmzury/food-delivery/services/delivery-service/internal/repository"
	deliverypb "github.com/mendmzury/food-delivery/proto/delivery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.Env)
	log := logger.WithService("delivery-service")

	log.Info().Msg("Starting delivery-service...")

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Postgres")
	}

	redisDB, err := database.NewRedisDB(cfg.RedisURI, cfg.RedisPassword)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisDB.Close()

	if err := db.AutoMigrate(&domain.DeliveryAssignment{}, &domain.DeliveryLocationHistory{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}

	assignmentRepo := repository.NewAssignmentRepository(db)
	trackingRepo := repository.NewTrackingRepository(db)
	deliveryService := application.NewDeliveryService(assignmentRepo, trackingRepo, redisDB, cfg)

	srv := server.NewDeliveryServer(deliveryService)
	grpcServer := grpc.NewServer()
	deliverypb.RegisterDeliveryServiceServer(grpcServer, srv)

	if cfg.IsDevelopment() {
		reflection.Register(grpcServer)
	}

	address := fmt.Sprintf(":%s", cfg.DeliveryServicePort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Str("address", address).Msg("Failed to listen")
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info().Msg("Shutting down delivery-service...")
		grpcServer.GracefulStop()
	}()

	log.Info().Str("address", address).Msg("delivery-service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve")
	}
}

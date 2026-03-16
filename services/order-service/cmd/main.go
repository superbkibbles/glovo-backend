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
	"github.com/mendmzury/food-delivery/services/order-service/internal/application"
	"github.com/mendmzury/food-delivery/services/order-service/internal/domain"
	"github.com/mendmzury/food-delivery/services/order-service/internal/grpc/server"
	"github.com/mendmzury/food-delivery/services/order-service/internal/repository"
	orderpb "github.com/mendmzury/food-delivery/proto/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.Env)
	log := logger.WithService("order-service")

	log.Info().Msg("Starting order-service...")

	db, err := database.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Postgres")
	}

	if err := db.AutoMigrate(&domain.Order{}, &domain.OrderItem{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}

	orderRepo := repository.NewOrderRepository(db)
	orderService := application.NewOrderService(orderRepo, cfg)

	srv := server.NewOrderServer(orderService)
	grpcServer := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(grpcServer, srv)

	if cfg.IsDevelopment() {
		reflection.Register(grpcServer)
	}

	address := fmt.Sprintf(":%s", cfg.OrderServicePort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Str("address", address).Msg("Failed to listen")
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info().Msg("Shutting down order-service...")
		grpcServer.GracefulStop()
	}()

	log.Info().Str("address", address).Msg("order-service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve")
	}
}

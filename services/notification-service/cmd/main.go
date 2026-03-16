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
	"github.com/mendmzury/food-delivery/services/notification-service/internal/grpc/server"
	notificationpb "github.com/mendmzury/food-delivery/proto/notification"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.Env)
	log := logger.WithService("notification-service")

	log.Info().Msg("Starting notification-service...")

	mongoDB, err := database.NewMongoDB(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongoDB.Close(context.Background())

	if err := mongoDB.CreateIndexes(context.Background()); err != nil {
		log.Warn().Err(err).Msg("Failed to create MongoDB indexes")
	}

	srv := server.NewNotificationServer()
	grpcServer := grpc.NewServer()
	notificationpb.RegisterNotificationServiceServer(grpcServer, srv)

	if cfg.IsDevelopment() {
		reflection.Register(grpcServer)
	}

	address := fmt.Sprintf(":%s", cfg.NotificationServicePort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Str("address", address).Msg("Failed to listen")
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info().Msg("Shutting down notification-service...")
		grpcServer.GracefulStop()
	}()

	log.Info().Str("address", address).Msg("notification-service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve")
	}
}

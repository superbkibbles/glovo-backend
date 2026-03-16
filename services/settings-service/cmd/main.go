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
	settingspb "github.com/mendmzury/food-delivery/proto/settings"
	"github.com/mendmzury/food-delivery/services/settings-service/internal/grpc/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.Env)
	log := logger.WithService("settings-service")

	log.Info().Msg("Starting settings-service...")

	mongoDB, err := database.NewMongoDB(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongoDB.Close(context.Background())

	if err := mongoDB.CreateIndexes(context.Background()); err != nil {
		log.Warn().Err(err).Msg("Failed to create MongoDB indexes")
	}

	srv := server.NewSettingsServer(mongoDB)
	grpcServer := grpc.NewServer()
	settingspb.RegisterSettingsServiceServer(grpcServer, srv)

	if cfg.IsDevelopment() {
		reflection.Register(grpcServer)
	}

	address := fmt.Sprintf(":%s", cfg.SettingsServicePort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Str("address", address).Msg("Failed to listen")
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info().Msg("Shutting down settings-service...")
		grpcServer.GracefulStop()
	}()

	log.Info().Str("address", address).Msg("settings-service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve")
	}
}

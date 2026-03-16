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
	"github.com/mendmzury/food-delivery/services/file-storage-service/internal/grpc/server"
	filepb "github.com/mendmzury/food-delivery/proto/file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.Env)
	log := logger.WithService("file-storage-service")

	log.Info().Msg("Starting file-storage-service...")

	mongoDB, err := database.NewMongoDB(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongoDB.Close(context.Background())

	srv := server.NewFileServer()
	grpcServer := grpc.NewServer()
	filepb.RegisterFileServiceServer(grpcServer, srv)

	if cfg.IsDevelopment() {
		reflection.Register(grpcServer)
	}

	address := fmt.Sprintf(":%s", cfg.FileStorageServicePort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Str("address", address).Msg("Failed to listen")
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Info().Msg("Shutting down file-storage-service...")
		grpcServer.GracefulStop()
	}()

	log.Info().Str("address", address).Msg("file-storage-service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve")
	}
}

package grpc

import (
	"context"

	"github.com/mendmzury/food-delivery/pkg/config"
	"github.com/mendmzury/food-delivery/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	authpb "github.com/mendmzury/food-delivery/proto/auth"
	userpb "github.com/mendmzury/food-delivery/proto/user"
	restaurantpb "github.com/mendmzury/food-delivery/proto/restaurant"
	orderpb "github.com/mendmzury/food-delivery/proto/order"
	deliverypb "github.com/mendmzury/food-delivery/proto/delivery"
	settingspb "github.com/mendmzury/food-delivery/proto/settings"
)

type Clients struct {
	Auth       authpb.AuthServiceClient
	User       userpb.UserServiceClient
	Role       userpb.RoleServiceClient
	Restaurant restaurantpb.RestaurantServiceClient
	Order      orderpb.OrderServiceClient
	Delivery   deliverypb.DeliveryServiceClient
	Settings   settingspb.SettingsServiceClient

	conns []*grpc.ClientConn
}

func NewClients(cfg *config.Config) (*Clients, error) {
	clients := &Clients{}

	authConn, err := connect(cfg.AuthServiceAddr)
	if err != nil {
		return nil, err
	}
	clients.Auth = authpb.NewAuthServiceClient(authConn)
	clients.conns = append(clients.conns, authConn)

	userConn, err := connect(cfg.UserServiceAddr)
	if err != nil {
		clients.Close()
		return nil, err
	}
	clients.User = userpb.NewUserServiceClient(userConn)
	clients.Role = userpb.NewRoleServiceClient(userConn)
	clients.conns = append(clients.conns, userConn)

	restConn, err := connect(cfg.RestaurantServiceAddr)
	if err != nil {
		clients.Close()
		return nil, err
	}
	clients.Restaurant = restaurantpb.NewRestaurantServiceClient(restConn)
	clients.conns = append(clients.conns, restConn)

	orderConn, err := connect(cfg.OrderServiceAddr)
	if err != nil {
		clients.Close()
		return nil, err
	}
	clients.Order = orderpb.NewOrderServiceClient(orderConn)
	clients.conns = append(clients.conns, orderConn)

	deliveryConn, err := connect(cfg.DeliveryServiceAddr)
	if err != nil {
		clients.Close()
		return nil, err
	}
	clients.Delivery = deliverypb.NewDeliveryServiceClient(deliveryConn)
	clients.conns = append(clients.conns, deliveryConn)

	settingsConn, err := connect(cfg.SettingsServiceAddr)
	if err != nil {
		clients.Close()
		return nil, err
	}
	clients.Settings = settingspb.NewSettingsServiceClient(settingsConn)
	clients.conns = append(clients.conns, settingsConn)

	logger.Info().Msg("Connected to gRPC services")
	return clients, nil
}

func (c *Clients) Close() error {
	for _, conn := range c.conns {
		if err := conn.Close(); err != nil {
			logger.Error().Err(err).Msg("Failed to close gRPC connection")
		}
	}
	return nil
}

func connect(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func WithAuth(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}

// WithDriverID sets the driver-id metadata header for order-service (accept, etc.).
func WithDriverID(ctx context.Context, driverID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "driver-id", driverID)
}

package grpc

import (
	"context"
	"time"

	pb "glovo-backend/proto/gen/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserGRPCClient struct {
	client pb.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserGRPCClient(address string) (*UserGRPCClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewUserServiceClient(conn)

	return &UserGRPCClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *UserGRPCClient) Close() error {
	return c.conn.Close()
}

func (c *UserGRPCClient) GetUser(ctx context.Context, userID string) (*pb.GetUserResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.GetUser(ctxWithTimeout, &pb.GetUserRequest{
		UserId: userID,
	})
}

func (c *UserGRPCClient) ValidateUser(ctx context.Context, userID string) (*pb.ValidateUserResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.ValidateUser(ctxWithTimeout, &pb.ValidateUserRequest{
		UserId: userID,
	})
}

func (c *UserGRPCClient) GetUsersBatch(ctx context.Context, userIDs []string) (*pb.GetUsersBatchResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return c.client.GetUsersBatch(ctxWithTimeout, &pb.GetUsersBatchRequest{
		UserIds: userIDs,
	})
}

func (c *UserGRPCClient) UpdateUserStatus(ctx context.Context, userID, status, reason string) (*pb.UpdateUserStatusResponse, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.UpdateUserStatus(ctxWithTimeout, &pb.UpdateUserStatusRequest{
		UserId: userID,
		Status: status,
		Reason: reason,
	})
}

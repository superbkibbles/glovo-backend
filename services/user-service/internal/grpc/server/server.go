package server

import (
	"context"

	"github.com/mendmzury/food-delivery/services/user-service/internal/application"
	commonpb "github.com/mendmzury/food-delivery/proto/common"
	userpb "github.com/mendmzury/food-delivery/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServer struct {
	userpb.UnimplementedUserServiceServer
	userpb.UnimplementedRoleServiceServer
	userService *application.UserService
}

func NewUserServer(userService *application.UserService) *UserServer {
	return &UserServer{
		userService: userService,
	}
}

func RegisterUserServer(grpcServer *grpc.Server, userService *application.UserService) {
	srv := NewUserServer(userService)
	userpb.RegisterUserServiceServer(grpcServer, srv)
	userpb.RegisterRoleServiceServer(grpcServer, srv)
}

func (s *UserServer) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	return s.userService.ListUsers(ctx, req)
}

func (s *UserServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.User, error) {
	return s.userService.GetUser(ctx, req)
}

func (s *UserServer) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.User, error) {
	return s.userService.CreateUser(ctx, req)
}

func (s *UserServer) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.User, error) {
	return s.userService.UpdateUser(ctx, req)
}

func (s *UserServer) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*commonpb.SuccessResponse, error) {
	return s.userService.DeleteUser(ctx, req)
}

func (s *UserServer) GetUserByUsername(ctx context.Context, req *userpb.GetUserByUsernameRequest) (*userpb.User, error) {
	return nil, status.Error(codes.Unimplemented, "GetUserByUsername not implemented")
}

func (s *UserServer) ListRoles(ctx context.Context, req *userpb.ListRolesRequest) (*userpb.ListRolesResponse, error) {
	return s.userService.ListRoles(ctx, req)
}

func (s *UserServer) GetRole(ctx context.Context, req *userpb.GetRoleRequest) (*userpb.Role, error) {
	return s.userService.GetRole(ctx, req)
}

func (s *UserServer) CreateRole(ctx context.Context, req *userpb.CreateRoleRequest) (*userpb.Role, error) {
	return nil, status.Error(codes.Unimplemented, "CreateRole not implemented")
}

func (s *UserServer) UpdateRole(ctx context.Context, req *userpb.UpdateRoleRequest) (*userpb.Role, error) {
	return nil, status.Error(codes.Unimplemented, "UpdateRole not implemented")
}

func (s *UserServer) DeleteRole(ctx context.Context, req *userpb.DeleteRoleRequest) (*commonpb.SuccessResponse, error) {
	return nil, status.Error(codes.Unimplemented, "DeleteRole not implemented")
}

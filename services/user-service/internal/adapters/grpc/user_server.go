package grpc

import (
	"context"

	"glovo-backend/services/user-service/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userGRPCServer struct {
	user.UnimplementedUserServiceServer
	userService domain.UserService
}

func NewUserGRPCServer(userService domain.UserService) user.UserServiceServer {
	return &userGRPCServer{
		userService: userService,
	}
}

func (s *userGRPCServer) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	if req.UserId == "" {
		return &user.GetUserResponse{
			Success: false,
			Error:   "user_id is required",
		}, status.Error(codes.InvalidArgument, "user_id is required")
	}

	domainUser, err := s.userService.GetUserByID(req.UserId)
	if err != nil {
		return &user.GetUserResponse{
			Success: false,
			Error:   err.Error(),
		}, status.Error(codes.NotFound, err.Error())
	}

	return &user.GetUserResponse{
		User:    convertDomainUserToProto(domainUser),
		Success: true,
	}, nil
}

func (s *userGRPCServer) ValidateUser(ctx context.Context, req *user.ValidateUserRequest) (*user.ValidateUserResponse, error) {
	if req.UserId == "" {
		return &user.ValidateUserResponse{
			IsValid: false,
		}, status.Error(codes.InvalidArgument, "user_id is required")
	}

	domainUser, err := s.userService.GetUserByID(req.UserId)
	if err != nil {
		return &user.ValidateUserResponse{
			IsValid:  false,
			IsActive: false,
		}, nil
	}

	return &user.ValidateUserResponse{
		IsValid:  true,
		IsActive: domainUser.IsActive,
		Role:     string(domainUser.Role),
	}, nil
}

func (s *userGRPCServer) GetUsersBatch(ctx context.Context, req *user.GetUsersBatchRequest) (*user.GetUsersBatchResponse, error) {
	if len(req.UserIds) == 0 {
		return &user.GetUsersBatchResponse{
			Users: []*user.User{},
		}, nil
	}

	var protoUsers []*user.User
	for _, userID := range req.UserIds {
		domainUser, err := s.userService.GetUserByID(userID)
		if err != nil {
			// Skip users that don't exist, don't fail the entire batch
			continue
		}
		protoUsers = append(protoUsers, convertDomainUserToProto(domainUser))
	}

	return &user.GetUsersBatchResponse{
		Users: protoUsers,
	}, nil
}

func (s *userGRPCServer) UpdateUserStatus(ctx context.Context, req *user.UpdateUserStatusRequest) (*user.UpdateUserStatusResponse, error) {
	if req.UserId == "" {
		return &user.UpdateUserStatusResponse{
			Success: false,
			Error:   "user_id is required",
		}, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// This would typically require admin authentication
	// For now, we'll implement a basic status update
	domainUser, err := s.userService.GetUserByID(req.UserId)
	if err != nil {
		return &user.UpdateUserStatusResponse{
			Success: false,
			Error:   err.Error(),
		}, status.Error(codes.NotFound, err.Error())
	}

	// Update user status based on the request
	switch req.Status {
	case "active":
		domainUser.IsActive = true
	case "inactive", "suspended":
		domainUser.IsActive = false
	default:
		return &user.UpdateUserStatusResponse{
			Success: false,
			Error:   "invalid status",
		}, status.Error(codes.InvalidArgument, "invalid status")
	}

	_, err = s.userService.UpdateProfile(domainUser.ID, domain.UpdateProfileRequest{
		// Profile update logic would go here
		// For now, we'll just mark the operation as successful
	})

	if err != nil {
		return &user.UpdateUserStatusResponse{
			Success: false,
			Error:   err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	return &user.UpdateUserStatusResponse{
		Success: true,
	}, nil
}

// Helper function to convert domain User to proto User
func convertDomainUserToProto(domainUser *domain.User) *user.User {
	return &user.User{
		Id:    domainUser.ID,
		Phone: domainUser.Phone,
		Email: domainUser.Email,
		Role:  string(domainUser.Role),
		Profile: &user.UserProfile{
			FirstName: domainUser.Profile.FirstName,
			LastName:  domainUser.Profile.LastName,
			Avatar:    domainUser.Profile.Avatar,
		},
		Status:    getStatusString(domainUser.IsActive),
		CreatedAt: domainUser.CreatedAt.Unix(),
		UpdatedAt: domainUser.UpdatedAt.Unix(),
	}
}

func getStatusString(isActive bool) string {
	if isActive {
		return "active"
	}
	return "inactive"
}

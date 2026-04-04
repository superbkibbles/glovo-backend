package server

import (
	"context"

	"github.com/mendmzury/food-delivery/pkg/middleware"
	authpb "github.com/mendmzury/food-delivery/proto/auth"
	"github.com/mendmzury/food-delivery/services/auth-service/internal/application"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	authService *application.AuthService
	permChecker *middleware.PermissionChecker
}

func NewAuthServer(authService *application.AuthService) *AuthServer {
	return &AuthServer{
		authService: authService,
		permChecker: middleware.NewPermissionChecker(),
	}
}

func RegisterAuthServer(grpcServer *grpc.Server, authService *application.AuthService) {
	authpb.RegisterAuthServiceServer(grpcServer, NewAuthServer(authService))
}

func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	result, err := s.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return toLoginResponse(result), nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	user, err := s.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return &authpb.ValidateTokenResponse{Valid: false}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid: true,
		User:  toUserInfoFromValidate(user),
	}, nil
}

func (s *AuthServer) ChangePassword(ctx context.Context, req *authpb.ChangePasswordRequest) (*authpb.ChangePasswordResponse, error) {
	claims, err := middleware.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.OldPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "old password is required")
	}
	if req.NewPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "new password is required")
	}

	if err := s.authService.ChangePassword(ctx, claims.UserID, req.OldPassword, req.NewPassword); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &authpb.ChangePasswordResponse{
		Success: true,
		Message: "Password changed successfully",
	}, nil
}

func (s *AuthServer) ResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest) (*authpb.ResetPasswordResponse, error) {
	if err := s.permChecker.RequireSuperadmin(ctx); err != nil {
		return nil, err
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.NewPassword == "" {
		return nil, status.Error(codes.InvalidArgument, "new password is required")
	}

	if err := s.authService.ResetPassword(ctx, req.UserId, req.NewPassword); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authpb.ResetPasswordResponse{
		Success: true,
		Message: "Password reset successfully",
	}, nil
}

func (s *AuthServer) ForgotPassword(ctx context.Context, req *authpb.ForgotPasswordRequest) (*authpb.ForgotPasswordResponse, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	phone, err := s.authService.ForgotPassword(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authpb.ForgotPasswordResponse{
		AdminPhone: phone,
		Message:    "Please contact the administrator to reset your password",
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	result, err := s.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &authpb.RefreshTokenResponse{
		Token:        result.Token,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		User:         toUserInfo(result.User),
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.LogoutResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	if err := s.authService.Logout(ctx, req.Token); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authpb.LogoutResponse{Success: true}, nil
}

func toLoginResponse(result *application.LoginResult) *authpb.LoginResponse {
	if result == nil {
		return nil
	}
	return &authpb.LoginResponse{
		Token:        result.Token,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		User:         toUserInfo(result.User),
		IsNewUser:    result.IsNewUser,
	}
}

func toUserInfo(u *application.UserInfo) *authpb.UserInfo {
	if u == nil {
		return nil
	}
	return &authpb.UserInfo{
		Id:           u.ID,
		Username:     u.Username,
		Name:         u.Name,
		RoleId:       u.RoleID,
		RoleName:     u.RoleName,
		IsSuperadmin: u.IsSuperadmin,
		Permissions:  u.Permissions,
		UserType:     u.UserType,
		Email:        u.Email,
		PhoneNumber:  u.PhoneNumber,
	}
}

func toUserInfoFromValidate(u *application.UserInfo) *authpb.UserInfo {
	if u == nil {
		return nil
	}
	return &authpb.UserInfo{
		Id:           u.ID,
		Username:     u.Username,
		CompanyId:    u.CompanyID,
		RoleId:       u.RoleID,
		IsSuperadmin: u.IsSuperadmin,
		Permissions:  u.Permissions,
		UserType:     u.UserType,
		Email:        u.Email,
		PhoneNumber:  u.PhoneNumber,
	}
}

func (s *AuthServer) SignUpWithPhone(ctx context.Context, req *authpb.SignUpWithPhoneRequest) (*authpb.LoginResponse, error) {
	if req.PhoneNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "phone_number is required")
	}
	if req.Otp == "" {
		return nil, status.Error(codes.InvalidArgument, "otp is required")
	}

	result, err := s.authService.SignUpWithPhone(ctx, req.PhoneNumber, req.Otp)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return toLoginResponse(result), nil
}

func (s *AuthServer) SignUpWithGoogle(ctx context.Context, req *authpb.SignUpWithGoogleRequest) (*authpb.LoginResponse, error) {
	if req.IdToken == "" {
		return nil, status.Error(codes.InvalidArgument, "id_token is required")
	}

	result, err := s.authService.SignUpWithGoogle(ctx, req.IdToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return toLoginResponse(result), nil
}

func (s *AuthServer) SignUpWithEmail(ctx context.Context, req *authpb.SignUpWithEmailRequest) (*authpb.LoginResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Otp == "" {
		return nil, status.Error(codes.InvalidArgument, "otp is required")
	}

	result, err := s.authService.SignUpWithEmail(ctx, req.Email, req.Otp, req.Name)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return toLoginResponse(result), nil
}

func (s *AuthServer) SignInWithPhone(ctx context.Context, req *authpb.SignInWithPhoneRequest) (*authpb.LoginResponse, error) {
	if req.PhoneNumber == "" {
		return nil, status.Error(codes.InvalidArgument, "phone_number is required")
	}
	if req.Otp == "" {
		return nil, status.Error(codes.InvalidArgument, "otp is required")
	}

	result, err := s.authService.SignInWithPhone(ctx, req.PhoneNumber, req.Otp)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return toLoginResponse(result), nil
}

func (s *AuthServer) SignInWithEmail(ctx context.Context, req *authpb.SignInWithEmailRequest) (*authpb.LoginResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Otp == "" {
		return nil, status.Error(codes.InvalidArgument, "otp is required")
	}

	result, err := s.authService.SignInWithEmail(ctx, req.Email, req.Otp)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return toLoginResponse(result), nil
}

func (s *AuthServer) SendOTP(ctx context.Context, req *authpb.SendOTPRequest) (*authpb.SendOTPResponse, error) {
	if req.PhoneNumber == "" && req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "phone_number or email is required")
	}
	if req.PhoneNumber != "" && req.Email != "" {
		return nil, status.Error(codes.InvalidArgument, "provide either phone_number or email, not both")
	}

	if err := s.authService.SendOTP(ctx, req.PhoneNumber, req.Email); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authpb.SendOTPResponse{
		Success: true,
		Message: "OTP sent successfully",
	}, nil
}

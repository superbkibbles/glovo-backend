package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/mendmzury/food-delivery/pkg/utils"
	"github.com/mendmzury/food-delivery/services/user-service/internal/domain"
	"github.com/mendmzury/food-delivery/services/user-service/internal/repository"
	userpb "github.com/mendmzury/food-delivery/proto/user"
	commonpb "github.com/mendmzury/food-delivery/proto/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService struct {
	userRepo *repository.UserRepository
	roleRepo *repository.RoleRepository
	pepper   string
}

func NewUserService(userRepo *repository.UserRepository, roleRepo *repository.RoleRepository, pepper string) *UserService {
	return &UserService{
		userRepo: userRepo,
		roleRepo: roleRepo,
		pepper:   pepper,
	}
}

func (s *UserService) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	page, pageSize := int64(1), int64(10)
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.PageSize > 0 {
			pageSize = req.Pagination.PageSize
		}
	}
	if pageSize > 100 {
		pageSize = 100
	}

	userType := ""
	if req.UserType != userpb.UserType_USER_TYPE_UNSPECIFIED {
		userType = userTypeToStr(req.UserType)
	}

	users, total, err := s.userRepo.List(ctx, req.Search, userType, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoUsers := make([]*userpb.User, len(users))
	for i, u := range users {
		role, _ := s.roleRepo.GetByID(ctx, u.RoleID)
		roleName := ""
		if role != nil {
			roleName = role.Name
		}
		protoUsers[i] = userToProto(u, roleName)
	}

	totalPages := int64(0)
	if total > 0 && pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return &userpb.ListUsersResponse{
		Users: protoUsers,
		Pagination: &commonpb.PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.User, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	role, _ := s.roleRepo.GetByID(ctx, user.RoleID)
	roleName := ""
	if role != nil {
		roleName = role.Name
	}
	return userToProto(user, roleName), nil
}

func (s *UserService) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.User, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role_id")
	}

	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil || role == nil {
		return nil, status.Error(codes.InvalidArgument, "role not found")
	}

	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if exists {
		return nil, status.Error(codes.AlreadyExists, "username already exists")
	}

	pwHasher := utils.NewPasswordHasher(s.pepper)
	salt, err := pwHasher.GenerateSalt()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	hash, err := pwHasher.HashPassword(req.Password, salt)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     req.Username,
		PasswordHash: hash,
		Salt:         salt,
		Name:         req.Name,
		PhoneNumber:  req.PhoneNumber,
		RoleID:       roleID,
		UserType:     userTypeToStr(req.UserType),
		Active:       true,
	}
	if user.UserType == "" {
		user.UserType = "customer"
	}

	if err = s.userRepo.Create(ctx, user); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return userToProto(user, role.Name), nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.User, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil || user == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = *req.PhoneNumber
	}
	if req.ProfilePhoto != nil {
		user.ProfilePhoto = *req.ProfilePhoto
	}
	if req.RoleId != nil {
		roleID, err := uuid.Parse(*req.RoleId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid role_id")
		}
		role, _ := s.roleRepo.GetByID(ctx, roleID)
		if role == nil {
			return nil, status.Error(codes.InvalidArgument, "role not found")
		}
		user.RoleID = roleID
	}
	if req.Active != nil {
		user.Active = *req.Active
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	role, _ := s.roleRepo.GetByID(ctx, user.RoleID)
	roleName := ""
	if role != nil {
		roleName = role.Name
	}
	return userToProto(user, roleName), nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*commonpb.SuccessResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &commonpb.SuccessResponse{Success: true}, nil
}

func (s *UserService) ListRoles(ctx context.Context, req *userpb.ListRolesRequest) (*userpb.ListRolesResponse, error) {
	roles, err := s.roleRepo.GetAll(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	protoRoles := make([]*userpb.Role, len(roles))
	for i, r := range roles {
		protoRoles[i] = roleToProto(r)
	}
	return &userpb.ListRolesResponse{Roles: protoRoles}, nil
}

func (s *UserService) GetRole(ctx context.Context, req *userpb.GetRoleRequest) (*userpb.Role, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid role id")
	}
	role, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if role == nil {
		return nil, status.Error(codes.NotFound, "role not found")
	}
	return roleToProto(role), nil
}

func userTypeToStr(t userpb.UserType) string {
	switch t {
	case userpb.UserType_USER_TYPE_CUSTOMER:
		return "customer"
	case userpb.UserType_USER_TYPE_RESTAURANT_OWNER:
		return "restaurant_owner"
	case userpb.UserType_USER_TYPE_DRIVER:
		return "driver"
	case userpb.UserType_USER_TYPE_ADMIN:
		return "admin"
	case userpb.UserType_USER_TYPE_UNSPECIFIED:
		fallthrough
	default:
		return ""
	}
}

func userToProto(u *domain.User, roleName string) *userpb.User {
	ut := userpb.UserType_USER_TYPE_UNSPECIFIED
	switch u.UserType {
	case "customer":
		ut = userpb.UserType_USER_TYPE_CUSTOMER
	case "restaurant_owner":
		ut = userpb.UserType_USER_TYPE_RESTAURANT_OWNER
	case "driver":
		ut = userpb.UserType_USER_TYPE_DRIVER
	case "admin":
		ut = userpb.UserType_USER_TYPE_ADMIN
	}
	return &userpb.User{
		Id:           u.ID.String(),
		Username:     u.Username,
		Name:         u.Name,
		PhoneNumber:  u.PhoneNumber,
		ProfilePhoto: u.ProfilePhoto,
		RoleId:       u.RoleID.String(),
		RoleName:     roleName,
		IsSuperadmin: u.IsSuperadmin,
		Active:       u.Active,
		UserType:     ut,
		CreatedAt:    timestamppb.New(u.CreatedAt),
		UpdatedAt:    timestamppb.New(u.UpdatedAt),
	}
}

func roleToProto(r *domain.Role) *userpb.Role {
	perms := []string(r.Permissions)
	if perms == nil {
		perms = []string{}
	}
	return &userpb.Role{
		Id:          r.ID.String(),
		Name:        r.Name,
		Icon:        r.Icon,
		Permissions: perms,
		CreatedAt:   timestamppb.New(r.CreatedAt),
		UpdatedAt:   timestamppb.New(r.UpdatedAt),
	}
}

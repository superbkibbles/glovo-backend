package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mendmzury/food-delivery/pkg/config"
	commonpb "github.com/mendmzury/food-delivery/proto/common"
	userpb "github.com/mendmzury/food-delivery/proto/user"
	"github.com/mendmzury/food-delivery/services/gateway/internal/grpc"
)

// ListUsers godoc
// @Summary List users
// @Tags users
// @Security BearerAuth
// @Param page query int false "Page"
// @Param page_size query int false "Page size"
// @Success 200 {object} object
// @Router /users [get]
func ListUsers(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		if requirePermission(c, "users.view") {
			return
		}
		page, pageSize := getPaginationParams(c)
		search := c.Query("search")
		userTypeStr := c.Query("user_type")

		userType := userpb.UserType_USER_TYPE_UNSPECIFIED
		switch userTypeStr {
		case "customer":
			userType = userpb.UserType_USER_TYPE_CUSTOMER
		case "restaurant_owner":
			userType = userpb.UserType_USER_TYPE_RESTAURANT_OWNER
		case "driver":
			userType = userpb.UserType_USER_TYPE_DRIVER
		case "admin":
			userType = userpb.UserType_USER_TYPE_ADMIN
		}

		ctx := getAuthContext(c)
		resp, err := clients.User.ListUsers(ctx, &userpb.ListUsersRequest{
			Pagination: &commonpb.PaginationRequest{
				Page:     page,
				PageSize: pageSize,
			},
			Search:   search,
			UserType: userType,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		paginatedResponse(c, resp.Users, resp.Pagination.Total, page, pageSize)
	}
}

func GetUser(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx := getAuthContext(c)
		user, err := clients.User.GetUser(ctx, &userpb.GetUserRequest{Id: id})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, user)
	}
}

func CreateUser(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		if requirePermission(c, "users.create") {
			return
		}
		var req struct {
			Username    string `json:"username" binding:"required"`
			Password    string `json:"password" binding:"required"`
			Name        string `json:"name"`
			PhoneNumber string `json:"phone_number"`
			RoleID      string `json:"role_id" binding:"required"`
			UserType    string `json:"user_type"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		ut := userpb.UserType_USER_TYPE_CUSTOMER
		switch req.UserType {
		case "restaurant_owner":
			ut = userpb.UserType_USER_TYPE_RESTAURANT_OWNER
		case "driver":
			ut = userpb.UserType_USER_TYPE_DRIVER
		case "admin":
			ut = userpb.UserType_USER_TYPE_ADMIN
		}
		ctx := getAuthContext(c)
		user, err := clients.User.CreateUser(ctx, &userpb.CreateUserRequest{
			Username:    req.Username,
			Password:    req.Password,
			Name:        req.Name,
			PhoneNumber: req.PhoneNumber,
			RoleId:      req.RoleID,
			UserType:    ut,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		createdResponse(c, user)
	}
}

func UpdateUser(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		if requirePermission(c, "users.edit") {
			return
		}
		id := c.Param("id")
		var req struct {
			Name         *string `json:"name"`
			PhoneNumber  *string `json:"phone_number"`
			ProfilePhoto *string `json:"profile_photo"`
			RoleID       *string `json:"role_id"`
			Active       *bool   `json:"active"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		updateReq := &userpb.UpdateUserRequest{Id: id}
		if req.Name != nil {
			updateReq.Name = req.Name
		}
		if req.PhoneNumber != nil {
			updateReq.PhoneNumber = req.PhoneNumber
		}
		if req.ProfilePhoto != nil {
			updateReq.ProfilePhoto = req.ProfilePhoto
		}
		if req.RoleID != nil {
			updateReq.RoleId = req.RoleID
		}
		if req.Active != nil {
			updateReq.Active = req.Active
		}
		ctx := getAuthContext(c)
		user, err := clients.User.UpdateUser(ctx, updateReq)
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, user)
	}
}

func DeleteUser(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		if requirePermission(c, "users.delete") {
			return
		}
		id := c.Param("id")
		ctx := getAuthContext(c)
		_, err := clients.User.DeleteUser(ctx, &userpb.DeleteUserRequest{Id: id})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, gin.H{"message": "user deleted successfully"})
	}
}

func ListRoles(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := getAuthContext(c)
		resp, err := clients.Role.ListRoles(ctx, &userpb.ListRolesRequest{})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, resp.Roles)
	}
}

func GetRole(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx := getAuthContext(c)
		role, err := clients.Role.GetRole(ctx, &userpb.GetRoleRequest{Id: id})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, role)
	}
}

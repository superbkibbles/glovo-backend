package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mendmzury/food-delivery/pkg/config"
	authpb "github.com/mendmzury/food-delivery/proto/auth"
	"github.com/mendmzury/food-delivery/services/gateway/internal/grpc"
)

func authUserResponse(u *authpb.UserInfo) gin.H {
	if u == nil {
		return gin.H{}
	}
	return gin.H{
		"id":            u.Id,
		"username":      u.Username,
		"name":          u.Name,
		"role_id":       u.RoleId,
		"role_name":     u.RoleName,
		"is_superadmin": u.IsSuperadmin,
		"permissions":   u.Permissions,
		"user_type":     u.UserType,
		"email":         u.Email,
		"phone_number":  u.PhoneNumber,
	}
}

func authSuccessResponse(c *gin.Context, resp *authpb.LoginResponse) {
	successResponse(c, gin.H{
		"token":         resp.Token,
		"refresh_token": resp.RefreshToken,
		"expires_in":    resp.ExpiresIn,
		"user":          authUserResponse(resp.User),
	})
}

// Login godoc
// @Summary Admin login
// @Description Login with username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{username=string,password=string} true "Login"
// @Success 200 {object} object
// @Router /auth/login [post]
func Login(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		ctx := c.Request.Context()
		resp, err := clients.Auth.Login(ctx, &authpb.LoginRequest{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		authSuccessResponse(c, resp)
	}
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{refresh_token=string} true "Refresh token"
// @Success 200 {object} object
// @Router /auth/refresh-token [post]
func RefreshToken(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		ctx := c.Request.Context()
		resp, err := clients.Auth.RefreshToken(ctx, &authpb.RefreshTokenRequest{
			RefreshToken: req.RefreshToken,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		authSuccessResponse(c, &authpb.LoginResponse{
			Token:        resp.Token,
			RefreshToken: resp.RefreshToken,
			ExpiresIn:    resp.ExpiresIn,
			User:         resp.User,
		})
	}
}

// GetCurrentUser godoc
// @Summary Get current user
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object
// @Router /auth/me [get]
func GetCurrentUser(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, exists := c.Get("token")
		if !exists {
			errorResponse(c, http.StatusUnauthorized, "missing token")
			return
		}
		ctx := grpc.WithAuth(c.Request.Context(), token.(string))
		resp, err := clients.Auth.ValidateToken(ctx, &authpb.ValidateTokenRequest{
			Token: token.(string),
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, authUserResponse(resp.User))
	}
}

// ChangePassword godoc
// @Summary Change password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{old_password=string,new_password=string} true "Change password"
// @Success 200 {object} object
// @Router /auth/change-password [post]
func ChangePassword(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OldPassword string `json:"old_password" binding:"required"`
			NewPassword string `json:"new_password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		token, _ := c.Get("token")
		ctx := grpc.WithAuth(c.Request.Context(), token.(string))
		resp, err := clients.Auth.ChangePassword(ctx, &authpb.ChangePasswordRequest{
			OldPassword: req.OldPassword,
			NewPassword: req.NewPassword,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, gin.H{"message": resp.Message})
	}
}

// Logout godoc
// @Summary Logout
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} object
// @Router /auth/logout [post]
func Logout(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, exists := c.Get("token")
		if !exists {
			errorResponse(c, http.StatusUnauthorized, "missing token")
			return
		}
		ctx := grpc.WithAuth(c.Request.Context(), token.(string))
		_, err := clients.Auth.Logout(ctx, &authpb.LogoutRequest{
			Token: token.(string),
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, gin.H{"message": "logged out successfully"})
	}
}

// SignUpWithPhone godoc
// @Summary Customer signup with phone
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{phone_number=string,otp=string} true "Signup"
// @Success 200 {object} object
// @Router /auth/signup/phone [post]
func SignUpWithPhone(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PhoneNumber string `json:"phone_number" binding:"required"`
			Otp         string `json:"otp" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		ctx := c.Request.Context()
		resp, err := clients.Auth.SignUpWithPhone(ctx, &authpb.SignUpWithPhoneRequest{
			PhoneNumber: req.PhoneNumber,
			Otp:         req.Otp,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		authSuccessResponse(c, resp)
	}
}

// SignUpWithGoogle godoc
// @Summary Customer signup with Google
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{id_token=string} true "Google ID token"
// @Success 200 {object} object
// @Router /auth/signup/google [post]
func SignUpWithGoogle(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			IdToken string `json:"id_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		ctx := c.Request.Context()
		resp, err := clients.Auth.SignUpWithGoogle(ctx, &authpb.SignUpWithGoogleRequest{
			IdToken: req.IdToken,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		authSuccessResponse(c, resp)
	}
}

// SignUpWithEmail godoc
// @Summary Customer signup with email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{email=string,otp=string,name=string} true "Signup"
// @Success 200 {object} object
// @Router /auth/signup/email [post]
func SignUpWithEmail(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required"`
			Otp   string `json:"otp" binding:"required"`
			Name  string `json:"name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		ctx := c.Request.Context()
		resp, err := clients.Auth.SignUpWithEmail(ctx, &authpb.SignUpWithEmailRequest{
			Email: req.Email,
			Otp:   req.Otp,
			Name:  req.Name,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		authSuccessResponse(c, resp)
	}
}

// SignInWithPhone godoc
// @Summary Customer signin with phone
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{phone_number=string,otp=string} true "Signin"
// @Success 200 {object} object
// @Router /auth/signin/phone [post]
func SignInWithPhone(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PhoneNumber string `json:"phone_number" binding:"required"`
			Otp         string `json:"otp" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		ctx := c.Request.Context()
		resp, err := clients.Auth.SignInWithPhone(ctx, &authpb.SignInWithPhoneRequest{
			PhoneNumber: req.PhoneNumber,
			Otp:         req.Otp,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		authSuccessResponse(c, resp)
	}
}

// SignInWithEmail godoc
// @Summary Customer signin with email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{email=string,otp=string} true "Signin"
// @Success 200 {object} object
// @Router /auth/signin/email [post]
func SignInWithEmail(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required"`
			Otp   string `json:"otp" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		ctx := c.Request.Context()
		resp, err := clients.Auth.SignInWithEmail(ctx, &authpb.SignInWithEmailRequest{
			Email: req.Email,
			Otp:   req.Otp,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		authSuccessResponse(c, resp)
	}
}

// SignInWithGoogle godoc
// @Summary Customer signin with Google
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{id_token=string} true "Google ID token"
// @Success 200 {object} object
// @Router /auth/signin/google [post]
func SignInWithGoogle(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return SignUpWithGoogle(cfg, clients)
}

// SendOTP godoc
// @Summary Send OTP to phone or email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{phone_number=string,email=string} true "Phone or email"
// @Success 200 {object} object
// @Router /auth/send-otp [post]
func SendOTP(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			PhoneNumber string `json:"phone_number"`
			Email       string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		if req.PhoneNumber == "" && req.Email == "" {
			errorResponse(c, http.StatusBadRequest, "phone_number or email is required")
			return
		}
		if req.PhoneNumber != "" && req.Email != "" {
			errorResponse(c, http.StatusBadRequest, "provide either phone_number or email, not both")
			return
		}
		ctx := c.Request.Context()
		resp, err := clients.Auth.SendOTP(ctx, &authpb.SendOTPRequest{
			PhoneNumber: req.PhoneNumber,
			Email:       req.Email,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, gin.H{"message": resp.Message})
	}
}

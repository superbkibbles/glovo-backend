package http

import (
	"net/http"
	"strconv"

	"glovo-backend/services/user-service/internal/domain"
	"glovo-backend/shared/auth"
	"glovo-backend/shared/middleware"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService domain.UserService
}

func NewUserHandler(userService domain.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) SetupRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		// Public routes (no auth required)
		v1.POST("/auth/send-otp", h.SendOTP)
		v1.POST("/auth/verify-otp", h.VerifyOTP)
		v1.POST("/auth/admin-login", h.AdminLogin)

		// Protected routes (auth required)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/profile", h.GetProfile)
			protected.PUT("/profile", h.UpdateProfile)
		}

		// Admin only routes
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.RequireRole(auth.RoleAdmin))
		{
			admin.GET("/users", h.ListUsers)
			admin.PUT("/users/:id/suspend", h.SuspendUser)
			admin.PUT("/users/:id/reactivate", h.ReactivateUser)
		}
	}
}

// SendOTP godoc
// @Summary Send OTP to phone number
// @Description Send a verification code to the provided phone number
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Phone number"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/send-otp [post]
func (h *UserHandler) SendOTP(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.SendOTP(req.PhoneNumber); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
}

// VerifyOTP godoc
// @Summary Verify OTP and login
// @Description Verify the OTP code and return JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.VerifyOTPRequest true "Phone number and OTP code"
// @Success 200 {object} domain.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/verify-otp [post]
func (h *UserHandler) VerifyOTP(c *gin.Context) {
	var req domain.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.userService.VerifyOTP(req.PhoneNumber, req.OTPCode)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// AdminLogin godoc
// @Summary Admin login with email and password
// @Description Authenticate admin user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.AdminLoginRequest true "Admin credentials"
// @Success 200 {object} domain.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/admin-login [post]
func (h *UserHandler) AdminLogin(c *gin.Context) {
	var req domain.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.userService.AdminLogin(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.User
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the profile of the authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UserProfile true "Profile data"
// @Success 200 {object} domain.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var profile domain.UserProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateProfile(userID, profile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListUsers godoc
// @Summary List all users (Admin only)
// @Description Get a paginated list of all users
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.User
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	users, err := h.userService.ListUsers(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// SuspendUser godoc
// @Summary Suspend a user (Admin only)
// @Description Suspend a user account
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/users/{id}/suspend [put]
func (h *UserHandler) SuspendUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	if err := h.userService.SuspendUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User suspended successfully"})
}

// ReactivateUser godoc
// @Summary Reactivate a user (Admin only)
// @Description Reactivate a suspended user account
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/users/{id}/reactivate [put]
func (h *UserHandler) ReactivateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	if err := h.userService.ReactivateUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User reactivated successfully"})
}

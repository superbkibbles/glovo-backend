package http

import (
	"net/http"
	"strconv"
	"time"

	"glovo-backend/services/admin-service/internal/domain"
	"glovo-backend/shared/auth"
	"glovo-backend/shared/middleware"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService domain.AdminService
}

func NewAdminHandler(adminService domain.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

func (h *AdminHandler) SetupRoutes(router *gin.RouterGroup) {
	// Public authentication routes
	auth := router.Group("/auth")
	{
		auth.POST("/login", h.login)
	}

	// Protected admin routes
	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	{
		// Profile management
		admin.GET("/profile", h.getProfile)

		// Admin management (super admin only)
		adminMgmt := admin.Group("/admins")
		{
			adminMgmt.POST("/", h.createAdmin)
			adminMgmt.GET("/", h.listAdmins)
			adminMgmt.PUT("/:id", h.updateAdmin)
			adminMgmt.DELETE("/:id", h.deleteAdmin)
		}

		// User management
		users := admin.Group("/users")
		{
			users.GET("/", h.getUsers)
			users.GET("/:id", h.getUser)
			users.PUT("/:id/status", h.updateUserStatus)
			users.POST("/:id/suspend", h.suspendUser)
			users.POST("/:id/reactivate", h.reactivateUser)
		}

		// Merchant management
		merchants := admin.Group("/merchants")
		{
			merchants.GET("/", h.getMerchants)
		}

		// Driver management
		drivers := admin.Group("/drivers")
		{
			drivers.GET("/", h.getDrivers)
		}

		// Platform analytics
		analytics := admin.Group("/analytics")
		{
			analytics.GET("/platform", h.getPlatformStats)
			analytics.GET("/revenue", h.getRevenueStats)
			analytics.GET("/orders", h.getOrderStats)
		}

		// System configuration
		config := admin.Group("/config")
		{
			config.GET("/", h.getAllSystemConfigs)
			config.GET("/:key", h.getSystemConfig)
			config.PUT("/:key", h.setSystemConfig)
		}

		// Audit logs
		audit := admin.Group("/audit")
		{
			audit.GET("/", h.getAuditLogs)
			audit.GET("/resource/:resource", h.getResourceAuditLogs)
		}
	}
}

// @Summary Admin login
// @Description Authenticate admin user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.AdminLoginRequest true "Login credentials"
// @Success 200 {object} domain.AdminLoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *AdminHandler) login(c *gin.Context) {
	var req domain.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.adminService.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get admin profile
// @Description Get current admin's profile
// @Tags admin
// @Produce json
// @Success 200 {object} domain.Admin
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /admin/profile [get]
func (h *AdminHandler) getProfile(c *gin.Context) {
	adminID := c.GetString("user_id")

	admin, err := h.adminService.GetProfile(adminID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin not found"})
		return
	}

	c.JSON(http.StatusOK, admin)
}

// @Summary Create admin
// @Description Create a new admin user
// @Tags admin
// @Accept json
// @Produce json
// @Param request body domain.CreateAdminRequest true "Admin creation request"
// @Success 201 {object} domain.Admin
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/admins [post]
func (h *AdminHandler) createAdmin(c *gin.Context) {
	adminID := c.GetString("user_id")

	var req domain.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	admin, err := h.adminService.CreateAdmin(adminID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, admin)
}

// @Summary List admins
// @Description Get list of admin users
// @Tags admin
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.Admin
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/admins [get]
func (h *AdminHandler) listAdmins(c *gin.Context) {
	adminID := c.GetString("user_id")

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	admins, err := h.adminService.ListAdmins(adminID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, admins)
}

// @Summary Update admin
// @Description Update an admin user
// @Tags admin
// @Accept json
// @Produce json
// @Param id path string true "Admin ID"
// @Param request body domain.UpdateAdminRequest true "Admin update request"
// @Success 200 {object} domain.Admin
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/admins/{id} [put]
func (h *AdminHandler) updateAdmin(c *gin.Context) {
	adminID := c.GetString("user_id")
	targetAdminID := c.Param("id")

	var req domain.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	admin, err := h.adminService.UpdateAdmin(adminID, targetAdminID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, admin)
}

// @Summary Delete admin
// @Description Delete an admin user
// @Tags admin
// @Produce json
// @Param id path string true "Admin ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/admins/{id} [delete]
func (h *AdminHandler) deleteAdmin(c *gin.Context) {
	adminID := c.GetString("user_id")
	targetAdminID := c.Param("id")

	err := h.adminService.DeleteAdmin(adminID, targetAdminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Admin deleted successfully"})
}

// @Summary Get users
// @Description Get list of platform users
// @Tags users
// @Produce json
// @Param search query string false "Search term"
// @Param role query string false "User role"
// @Param status query string false "User status"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.UserInfo
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/users [get]
func (h *AdminHandler) getUsers(c *gin.Context) {
	req := domain.UserSearchRequest{
		Query:  c.Query("search"),
		Role:   auth.UserRole(c.Query("role")),
		Status: domain.UserStatus(c.Query("status")),
	}

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	req.Limit, _ = strconv.Atoi(limitStr)
	req.Offset, _ = strconv.Atoi(offsetStr)

	users, err := h.adminService.GetUsers(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// @Summary Get user
// @Description Get specific user details
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} domain.UserInfo
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /admin/users/{id} [get]
func (h *AdminHandler) getUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.adminService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Summary Update user status
// @Description Update user account status
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body domain.UpdateUserStatusRequest true "Status update request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/users/{id}/status [put]
func (h *AdminHandler) updateUserStatus(c *gin.Context) {
	adminID := c.GetString("user_id")
	userID := c.Param("id")

	var req domain.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.adminService.UpdateUserStatus(adminID, userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User status updated successfully"})
}

// @Summary Suspend user
// @Description Suspend a user account
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body struct{Reason string `json:"reason"`} true "Suspension reason"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/users/{id}/suspend [post]
func (h *AdminHandler) suspendUser(c *gin.Context) {
	adminID := c.GetString("user_id")
	userID := c.Param("id")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.adminService.SuspendUser(adminID, userID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User suspended successfully"})
}

// @Summary Reactivate user
// @Description Reactivate a suspended user account
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/users/{id}/reactivate [post]
func (h *AdminHandler) reactivateUser(c *gin.Context) {
	adminID := c.GetString("user_id")
	userID := c.Param("id")

	err := h.adminService.ReactivateUser(adminID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User reactivated successfully"})
}

// @Summary Get merchants
// @Description Get list of merchants
// @Tags merchants
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.MerchantInfo
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/merchants [get]
func (h *AdminHandler) getMerchants(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	merchants, err := h.adminService.GetMerchants(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, merchants)
}

// @Summary Get drivers
// @Description Get list of drivers
// @Tags drivers
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.DriverInfo
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/drivers [get]
func (h *AdminHandler) getDrivers(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	drivers, err := h.adminService.GetDrivers(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, drivers)
}

// @Summary Get platform stats
// @Description Get platform analytics and statistics
// @Tags analytics
// @Produce json
// @Success 200 {object} domain.PlatformStats
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/analytics/platform [get]
func (h *AdminHandler) getPlatformStats(c *gin.Context) {
	stats, err := h.adminService.GetPlatformStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary Get revenue stats
// @Description Get revenue statistics for a date range
// @Tags analytics
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {array} domain.RevenueStat
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/analytics/revenue [get]
func (h *AdminHandler) getRevenueStats(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (YYYY-MM-DD)"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (YYYY-MM-DD)"})
		return
	}

	stats, err := h.adminService.GetRevenueStats(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary Get order stats
// @Description Get order statistics for a date range
// @Tags analytics
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]int
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/analytics/orders [get]
func (h *AdminHandler) getOrderStats(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (YYYY-MM-DD)"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (YYYY-MM-DD)"})
		return
	}

	stats, err := h.adminService.GetOrderStats(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary Get system config
// @Description Get system configuration value
// @Tags config
// @Produce json
// @Param key path string true "Configuration key"
// @Success 200 {object} domain.SystemConfig
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /admin/config/{key} [get]
func (h *AdminHandler) getSystemConfig(c *gin.Context) {
	key := c.Param("key")

	config, err := h.adminService.GetSystemConfig(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// @Summary Set system config
// @Description Set system configuration value
// @Tags config
// @Accept json
// @Produce json
// @Param key path string true "Configuration key"
// @Param request body domain.SystemConfigRequest true "Configuration request"
// @Success 200 {object} domain.SystemConfig
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/config/{key} [put]
func (h *AdminHandler) setSystemConfig(c *gin.Context) {
	adminID := c.GetString("user_id")
	key := c.Param("key")

	var req domain.SystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the key from the URL parameter
	req.Key = key

	config, err := h.adminService.SetSystemConfig(adminID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// @Summary Get all system configs
// @Description Get all system configuration values
// @Tags config
// @Produce json
// @Success 200 {array} domain.SystemConfig
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/config [get]
func (h *AdminHandler) getAllSystemConfigs(c *gin.Context) {
	configs, err := h.adminService.GetAllSystemConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, configs)
}

// @Summary Get audit logs
// @Description Get audit logs
// @Tags audit
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.AuditLog
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/audit [get]
func (h *AdminHandler) getAuditLogs(c *gin.Context) {
	adminID := c.GetString("user_id")

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	logs, err := h.adminService.GetAuditLogs(adminID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// @Summary Get resource audit logs
// @Description Get audit logs for a specific resource
// @Tags audit
// @Produce json
// @Param resource path string true "Resource type"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.AuditLog
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /admin/audit/resource/{resource} [get]
func (h *AdminHandler) getResourceAuditLogs(c *gin.Context) {
	resource := c.Param("resource")

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	logs, err := h.adminService.GetResourceAuditLogs(resource, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

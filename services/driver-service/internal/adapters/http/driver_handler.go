package http

import (
	"net/http"
	"strconv"
	"time"

	"glovo-backend/services/driver-service/internal/domain"
	"glovo-backend/shared/auth"
	"glovo-backend/shared/middleware"

	"github.com/gin-gonic/gin"
)

type DriverHandler struct {
	driverService domain.DriverService
}

func NewDriverHandler(driverService domain.DriverService) *DriverHandler {
	return &DriverHandler{
		driverService: driverService,
	}
}

func (h *DriverHandler) SetupRoutes(router *gin.RouterGroup) {
	// Public driver registration routes
	public := router.Group("/drivers")
	{
		public.POST("/register", h.registerDriver)
	}

	// Driver profile management (authenticated drivers)
	profile := router.Group("/driver/profile")
	profile.Use(middleware.AuthMiddleware())
	profile.Use(middleware.RequireRoles([]auth.UserRole{auth.RoleDriver, auth.RoleAdmin}))
	{
		profile.GET("/", h.getProfile)
		profile.PUT("/", h.updateProfile)
		profile.PUT("/status", h.updateStatus)
		profile.PUT("/location", h.updateLocation)

		// Document management
		profile.POST("/documents", h.uploadDocument)
		profile.GET("/documents", h.getDocuments)

		// Earnings
		profile.GET("/earnings", h.getEarningsReport)
	}

	// Admin driver management
	admin := router.Group("/admin/drivers")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RequireRole(auth.RoleAdmin))
	{
		admin.GET("/", h.searchDrivers)
		admin.GET("/:id", h.getDriver)
		admin.GET("/available", h.getAvailableDrivers)
		admin.PUT("/documents/:id/approve", h.approveDocument)
		admin.PUT("/documents/:id/reject", h.rejectDocument)
		admin.PUT("/:id/performance", h.updatePerformance)
	}
}

// @Summary Register as driver
// @Description Register a new driver profile
// @Tags drivers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.RegisterDriverRequest true "Driver registration data"
// @Success 201 {object} domain.Driver
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/drivers/register [post]
func (h *DriverHandler) registerDriver(c *gin.Context) {
	var req domain.RegisterDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	driver, err := h.driverService.RegisterDriver(userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, driver)
}

// @Summary Get driver profile
// @Description Get the authenticated driver's profile
// @Tags drivers
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Driver
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/profile [get]
func (h *DriverHandler) getProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	driver, err := h.driverService.GetDriverByUser(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, driver)
}

// @Summary Update driver profile
// @Description Update the authenticated driver's profile information
// @Tags drivers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UpdateDriverProfileRequest true "Profile update data"
// @Success 200 {object} domain.Driver
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/profile [put]
func (h *DriverHandler) updateProfile(c *gin.Context) {
	var req domain.UpdateDriverProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	driverID := c.Query("driver_id")

	driver, err := h.driverService.UpdateProfile(driverID, userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, driver)
}

// @Summary Update driver status
// @Description Update driver availability status
// @Tags drivers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.DriverStatus true "Status update"
// @Success 200 {object} domain.Driver
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/profile/status [put]
func (h *DriverHandler) updateStatus(c *gin.Context) {
	var req struct {
		Status domain.DriverStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	driverID := c.Query("driver_id")

	driver, err := h.driverService.UpdateStatus(driverID, userID.(string), req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, driver)
}

// @Summary Update driver location
// @Description Update driver's current location
// @Tags drivers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UpdateLocationRequest true "Location data"
// @Success 200 {object} domain.Driver
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/profile/location [put]
func (h *DriverHandler) updateLocation(c *gin.Context) {
	var req domain.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	driverID := c.Query("driver_id")

	driver, err := h.driverService.UpdateLocation(driverID, userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, driver)
}

// @Summary Upload driver document
// @Description Upload a required document for driver verification
// @Tags drivers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UploadDocumentRequest true "Document data"
// @Success 201 {object} domain.DriverDocument
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/profile/documents [post]
func (h *DriverHandler) uploadDocument(c *gin.Context) {
	var req domain.UploadDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	driverID := c.Query("driver_id")

	document, err := h.driverService.UploadDocument(driverID, userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, document)
}

// @Summary Get driver documents
// @Description Get all documents for the authenticated driver
// @Tags drivers
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.DriverDocument
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/profile/documents [get]
func (h *DriverHandler) getDocuments(c *gin.Context) {
	userID, _ := c.Get("user_id")
	driverID := c.Query("driver_id")

	documents, err := h.driverService.GetDocuments(driverID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, documents)
}

// @Summary Get earnings report
// @Description Get driver earnings report for a time period
// @Tags drivers
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {array} domain.EarningsReport
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/profile/earnings [get]
func (h *DriverHandler) getEarningsReport(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
		return
	}

	userID, _ := c.Get("user_id")
	driverID := c.Query("driver_id")

	req := domain.EarningsReportRequest{
		StartDate: startDate,
		EndDate:   endDate,
	}

	report, err := h.driverService.GetEarningsReport(driverID, userID.(string), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// Admin endpoints

// @Summary Search drivers
// @Description Search for drivers with filters (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param status query string false "Driver status filter"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} domain.Driver
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/drivers [get]
func (h *DriverHandler) searchDrivers(c *gin.Context) {
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	req := domain.DriverSearchRequest{
		Status: domain.DriverStatus(status),
		Limit:  limit,
		Offset: offset,
	}

	drivers, err := h.driverService.SearchDrivers(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, drivers)
}

// @Summary Get driver by ID
// @Description Get driver details by ID (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Driver ID"
// @Success 200 {object} domain.Driver
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/drivers/{id} [get]
func (h *DriverHandler) getDriver(c *gin.Context) {
	driverID := c.Param("id")

	driver, err := h.driverService.GetDriver(driverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, driver)
}

// @Summary Get available drivers
// @Description Get drivers available for delivery assignment
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param latitude query number true "Latitude"
// @Param longitude query number true "Longitude"
// @Param radius query number false "Search radius in km"
// @Success 200 {array} domain.Driver
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/drivers/available [get]
func (h *DriverHandler) getAvailableDrivers(c *gin.Context) {
	latStr := c.Query("latitude")
	lngStr := c.Query("longitude")
	radiusStr := c.DefaultQuery("radius", "10")

	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "latitude and longitude are required"})
		return
	}

	latitude, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude"})
		return
	}

	longitude, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude"})
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid radius"})
		return
	}

	drivers, err := h.driverService.GetAvailableDrivers(latitude, longitude, radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, drivers)
}

// @Summary Approve driver document
// @Description Approve a driver document (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Document ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/drivers/documents/{id}/approve [put]
func (h *DriverHandler) approveDocument(c *gin.Context) {
	documentID := c.Param("id")
	adminID, _ := c.Get("user_id")

	err := h.driverService.ApproveDocument(documentID, adminID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document approved successfully"})
}

// @Summary Reject driver document
// @Description Reject a driver document with reason (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Document ID"
// @Param request body map[string]string true "Rejection reason"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/drivers/documents/{id}/reject [put]
func (h *DriverHandler) rejectDocument(c *gin.Context) {
	documentID := c.Param("id")
	adminID, _ := c.Get("user_id")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.driverService.RejectDocument(documentID, adminID.(string), req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document rejected successfully"})
}

// @Summary Update driver performance
// @Description Update driver performance metrics (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Driver ID"
// @Param request body domain.PerformanceStats true "Performance data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/drivers/{id}/performance [put]
func (h *DriverHandler) updatePerformance(c *gin.Context) {
	driverID := c.Param("id")

	var stats domain.PerformanceStats
	if err := c.ShouldBindJSON(&stats); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.driverService.UpdatePerformance(driverID, stats)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Performance updated successfully"})
}

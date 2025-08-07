package http

import (
	"net/http"
	"strconv"

	"glovo-backend/services/delivery-service/internal/domain"
	"glovo-backend/shared/auth"
	"glovo-backend/shared/middleware"

	"github.com/gin-gonic/gin"
)

type DeliveryHandler struct {
	deliveryService domain.DeliveryService
}

func NewDeliveryHandler(deliveryService domain.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{
		deliveryService: deliveryService,
	}
}

func (h *DeliveryHandler) SetupRoutes(router *gin.RouterGroup) {
	// Public delivery endpoints (for order service integration)
	public := router.Group("/deliveries")
	{
		public.POST("/create", h.createDelivery)
		public.GET("/:id/status", h.getDeliveryStatus)
	}

	// Driver delivery management
	driver := router.Group("/driver/deliveries")
	driver.Use(middleware.AuthMiddleware())
	driver.Use(middleware.RequireRole(auth.RoleDriver))
	{
		driver.GET("/available", h.getAvailableDeliveries)
		driver.POST("/:id/accept", h.acceptDelivery)
		driver.POST("/:id/pickup", h.markPickedUp)
		driver.POST("/:id/deliver", h.markDelivered)
		driver.POST("/:id/complete", h.completeDelivery)
		driver.POST("/:id/issue", h.reportIssue)
		driver.GET("/active", h.getActiveDeliveries)
		driver.GET("/history", h.getDeliveryHistory)
	}

	// Customer delivery tracking
	customer := router.Group("/customer/deliveries")
	customer.Use(middleware.AuthMiddleware())
	customer.Use(middleware.RequireRole(auth.RoleCustomer))
	{
		customer.GET("/:id/track", h.trackDelivery)
		customer.GET("/", h.getCustomerDeliveries)
	}

	// Admin delivery management
	admin := router.Group("/admin/deliveries")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RequireRole(auth.RoleAdmin))
	{
		admin.GET("/", h.searchDeliveries)
		admin.GET("/:id", h.getDelivery)
		admin.PUT("/:id/assign", h.assignDelivery)
		admin.PUT("/:id/reassign", h.reassignDelivery)
		admin.PUT("/:id/cancel", h.cancelDelivery)
		admin.GET("/metrics", h.getDeliveryMetrics)
		admin.GET("/drivers/:id/performance", h.getDriverPerformance)
		admin.GET("/drivers/rankings", h.getDriverRankings)
		admin.GET("/system/stats", h.getSystemStats)
	}
}

// @Summary Create delivery
// @Description Create a new delivery assignment
// @Tags deliveries
// @Accept json
// @Produce json
// @Param request body domain.CreateDeliveryRequest true "Delivery data"
// @Success 201 {object} domain.Delivery
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/deliveries/create [post]
func (h *DeliveryHandler) createDelivery(c *gin.Context) {
	var req domain.CreateDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	delivery, err := h.deliveryService.CreateDelivery(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, delivery)
}

// @Summary Get delivery status
// @Description Get the current status of a delivery
// @Tags deliveries
// @Produce json
// @Param id path string true "Delivery ID"
// @Success 200 {object} domain.DeliveryStatusResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/deliveries/{id}/status [get]
func (h *DeliveryHandler) getDeliveryStatus(c *gin.Context) {
	deliveryID := c.Param("id")

	delivery, err := h.deliveryService.GetDelivery(deliveryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": delivery.Delivery.Status})
}

// Driver endpoints

// @Summary Get available deliveries
// @Description Get deliveries available for the authenticated driver
// @Tags driver
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results"
// @Success 200 {array} domain.Delivery
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/deliveries/available [get]
func (h *DeliveryHandler) getAvailableDeliveries(c *gin.Context) {
	userID, _ := c.Get("user_id")

	assignments, err := h.deliveryService.GetPendingAssignments(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, assignments)
}

// @Summary Accept delivery
// @Description Accept a delivery assignment
// @Tags driver
// @Produce json
// @Security BearerAuth
// @Param id path string true "Delivery ID"
// @Success 200 {object} domain.Delivery
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/deliveries/{id}/accept [post]
func (h *DeliveryHandler) acceptDelivery(c *gin.Context) {
	deliveryID := c.Param("id")
	driverID, _ := c.Get("user_id")

	req := domain.DriverResponseRequest{
		DeliveryID: deliveryID,
		Accept:     true,
	}

	err := h.deliveryService.RespondToAssignment(driverID.(string), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get updated delivery to return
	delivery, err := h.deliveryService.GetDelivery(deliveryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, delivery)
}

// @Summary Mark as picked up
// @Description Mark delivery as picked up from merchant
// @Tags driver
// @Produce json
// @Security BearerAuth
// @Param id path string true "Delivery ID"
// @Success 200 {object} domain.Delivery
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/deliveries/{id}/pickup [post]
func (h *DeliveryHandler) markPickedUp(c *gin.Context) {
	deliveryID := c.Param("id")
	driverID, _ := c.Get("user_id")

	delivery, err := h.deliveryService.PickupOrder(deliveryID, driverID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, delivery)
}

// @Summary Mark as delivered
// @Description Mark delivery as delivered to customer
// @Tags driver
// @Produce json
// @Security BearerAuth
// @Param id path string true "Delivery ID"
// @Success 200 {object} domain.Delivery
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/deliveries/{id}/deliver [post]
func (h *DeliveryHandler) markDelivered(c *gin.Context) {
	deliveryID := c.Param("id")
	driverID, _ := c.Get("user_id")

	delivery, err := h.deliveryService.CompleteDelivery(deliveryID, driverID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, delivery)
}

// @Summary Complete delivery
// @Description Complete the delivery process
// @Tags driver
// @Produce json
// @Security BearerAuth
// @Param id path string true "Delivery ID"
// @Success 200 {object} domain.DeliveryResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/deliveries/{id}/complete [post]
func (h *DeliveryHandler) completeDelivery(c *gin.Context) {
	deliveryID := c.Param("id")
	driverID, _ := c.Get("user_id")

	response, err := h.deliveryService.CompleteDelivery(deliveryID, driverID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Report delivery issue
// @Description Report an issue with the delivery
// @Tags driver
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Delivery ID"
// @Param request body map[string]string true "Issue description"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/deliveries/{id}/issue [post]
func (h *DeliveryHandler) reportIssue(c *gin.Context) {
	deliveryID := c.Param("id")
	driverID, _ := c.Get("user_id")

	var req struct {
		Issue string `json:"issue" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.deliveryService.ReportIssue(deliveryID, driverID.(string), req.Issue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Issue reported successfully"})
}

// @Summary Get active deliveries
// @Description Get active deliveries for the authenticated driver
// @Tags driver
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.Delivery
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/deliveries/active [get]
func (h *DeliveryHandler) getActiveDeliveries(c *gin.Context) {
	driverID, _ := c.Get("user_id")

	deliveries, err := h.deliveryService.GetDriverAssignments(driverID.(string), 50, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deliveries)
}

// @Summary Get delivery history
// @Description Get delivery history for the authenticated driver
// @Tags driver
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} domain.Delivery
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/deliveries/history [get]
func (h *DeliveryHandler) getDeliveryHistory(c *gin.Context) {
	driverID, _ := c.Get("user_id")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	deliveries, err := h.deliveryService.GetDriverAssignments(driverID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deliveries)
}

// Customer endpoints

// @Summary Track delivery
// @Description Track delivery progress for customer
// @Tags customer
// @Produce json
// @Security BearerAuth
// @Param id path string true "Delivery ID"
// @Success 200 {object} domain.TrackingInfo
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/customer/deliveries/{id}/track [get]
func (h *DeliveryHandler) trackDelivery(c *gin.Context) {
	deliveryID := c.Param("id")
	_, _ = c.Get("user_id") // Customer validation can be added here if needed

	delivery, err := h.deliveryService.GetDelivery(deliveryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Return tracking info from delivery data
	trackingInfo := map[string]interface{}{
		"delivery_id":      delivery.Delivery.ID,
		"status":           delivery.Delivery.Status,
		"pickup_address":   delivery.Delivery.PickupAddress,
		"delivery_address": delivery.Delivery.DeliveryAddress,
		"estimated_time":   delivery.Delivery.EstimatedTime,
		"driver_id":        delivery.Delivery.DriverID,
		"picked_up_at":     delivery.Delivery.PickedUpAt,
		"delivered_at":     delivery.Delivery.DeliveredAt,
	}

	c.JSON(http.StatusOK, trackingInfo)
}

// @Summary Get customer deliveries
// @Description Get deliveries for the authenticated customer
// @Tags customer
// @Produce json
// @Security BearerAuth
// @Param status query string false "Delivery status filter"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} domain.Delivery
// @Failure 500 {object} map[string]string
// @Router /api/v1/customer/deliveries [get]
func (h *DeliveryHandler) getCustomerDeliveries(c *gin.Context) {
	_, _ = c.Get("user_id") // Customer filtering can be added here if needed
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	var deliveryStatus domain.DeliveryStatus
	if status != "" {
		deliveryStatus = domain.DeliveryStatus(status)
	}

	req := domain.DeliverySearchRequest{
		Status: deliveryStatus,
		Limit:  limit,
		Offset: offset,
	}

	deliveries, err := h.deliveryService.SearchDeliveries(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deliveries)
}

// Admin endpoints

// @Summary Search deliveries
// @Description Search deliveries with filters (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param status query string false "Delivery status filter"
// @Param driver_id query string false "Driver ID filter"
// @Param customer_id query string false "Customer ID filter"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} domain.Delivery
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/deliveries [get]
func (h *DeliveryHandler) searchDeliveries(c *gin.Context) {
	status := c.Query("status")
	driverID := c.Query("driver_id")
	_ = c.Query("customer_id") // Customer filtering not supported in current domain model
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	var deliveryStatus domain.DeliveryStatus
	if status != "" {
		deliveryStatus = domain.DeliveryStatus(status)
	}

	req := domain.DeliverySearchRequest{
		Status:   deliveryStatus,
		DriverID: driverID,
		Limit:    limit,
		Offset:   offset,
	}

	deliveries, err := h.deliveryService.SearchDeliveries(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, deliveries)
}

// @Summary Get delivery details
// @Description Get detailed delivery information (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Delivery ID"
// @Success 200 {object} domain.Delivery
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/deliveries/{id} [get]
func (h *DeliveryHandler) getDelivery(c *gin.Context) {
	deliveryID := c.Param("id")

	delivery, err := h.deliveryService.GetDelivery(deliveryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, delivery)
}

// @Summary Assign delivery
// @Description Assign delivery to a specific driver (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Delivery ID"
// @Param request body map[string]string true "Driver assignment"
// @Success 200 {object} domain.Delivery
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/deliveries/{id}/assign [put]
func (h *DeliveryHandler) assignDelivery(c *gin.Context) {
	deliveryID := c.Param("id")
	adminID, _ := c.Get("user_id")

	var req struct {
		DriverID string `json:"driver_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	assignReq := domain.AssignDriverRequest{
		DeliveryID: deliveryID,
		DriverID:   req.DriverID,
		Type:       domain.AssignmentManual,
	}

	delivery, err := h.deliveryService.ManualAssignDriver(assignReq, adminID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, delivery)
}

// @Summary Reassign delivery
// @Description Reassign delivery to a different driver (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Delivery ID"
// @Param request body map[string]string true "New driver assignment"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/deliveries/{id}/reassign [put]
func (h *DeliveryHandler) reassignDelivery(c *gin.Context) {
	deliveryID := c.Param("id")
	adminID, _ := c.Get("user_id")

	var req struct {
		NewDriverID string `json:"new_driver_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.deliveryService.ReassignDelivery(deliveryID, req.NewDriverID, adminID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery reassigned successfully"})
}

// @Summary Cancel delivery
// @Description Cancel a delivery (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Delivery ID"
// @Param request body map[string]string true "Cancellation reason"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/deliveries/{id}/cancel [put]
func (h *DeliveryHandler) cancelDelivery(c *gin.Context) {
	deliveryID := c.Param("id")
	adminID, _ := c.Get("user_id")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.deliveryService.CancelDelivery(deliveryID, req.Reason, adminID.(string), auth.RoleAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery cancelled successfully"})
}

// @Summary Get delivery metrics
// @Description Get delivery performance metrics (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.DeliveryMetrics
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/deliveries/metrics [get]
func (h *DeliveryHandler) getDeliveryMetrics(c *gin.Context) {
	metrics, err := h.deliveryService.GetDeliveryMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// @Summary Get driver performance
// @Description Get performance metrics for a specific driver (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Driver ID"
// @Success 200 {object} domain.DriverPerformance
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/deliveries/drivers/{id}/performance [get]
func (h *DeliveryHandler) getDriverPerformance(c *gin.Context) {
	driverID := c.Param("id")

	performance, err := h.deliveryService.GetDriverPerformance(driverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, performance)
}

// @Summary Get driver rankings
// @Description Get driver performance rankings (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.DriverPerformance
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/deliveries/drivers/rankings [get]
func (h *DeliveryHandler) getDriverRankings(c *gin.Context) {
	rankings, err := h.deliveryService.GetDriverRankings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rankings)
}

// @Summary Get system statistics
// @Description Get system-wide delivery statistics (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.DeliveryMetrics
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/deliveries/system/stats [get]
func (h *DeliveryHandler) getSystemStats(c *gin.Context) {
	stats, err := h.deliveryService.GetSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

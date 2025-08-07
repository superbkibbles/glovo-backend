package http

import (
	"net/http"
	"strconv"
	"time"

	"glovo-backend/services/analytics-service/internal/domain"
	"glovo-backend/shared/auth"
	"glovo-backend/shared/middleware"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	analyticsService domain.AnalyticsService
}

func NewAnalyticsHandler(analyticsService domain.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

func (h *AnalyticsHandler) SetupRoutes(router *gin.RouterGroup) {
	// Public analytics endpoints (for internal service integration)
	public := router.Group("/analytics")
	{
		public.POST("/track-event", h.trackEvent)
		public.POST("/track-order", h.trackOrderEvent)
		public.POST("/track-delivery", h.trackDeliveryEvent)
	}

	// Admin analytics dashboard
	admin := router.Group("/admin/analytics")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.RequireRole(auth.RoleAdmin))
	{
		// Platform overview
		admin.GET("/dashboard", h.getDashboard)
		admin.GET("/platform-stats", h.getPlatformStats)

		// Revenue analytics
		admin.GET("/revenue/overview", h.getRevenueOverview)
		admin.GET("/revenue/trends", h.getRevenueTrends)
		admin.GET("/revenue/by-category", h.getRevenueByCategory)
		admin.GET("/revenue/by-location", h.getRevenueByLocation)

		// Order analytics
		admin.GET("/orders/overview", h.getOrderOverview)
		admin.GET("/orders/trends", h.getOrderTrends)
		admin.GET("/orders/by-status", h.getOrdersByStatus)
		admin.GET("/orders/peak-hours", h.getPeakHours)

		// User analytics
		admin.GET("/users/overview", h.getUserOverview)
		admin.GET("/users/activity", h.getUserActivity)
		admin.GET("/users/retention", h.getUserRetention)
		admin.GET("/users/demographics", h.getUserDemographics)

		// Delivery analytics
		admin.GET("/deliveries/overview", h.getDeliveryOverview)
		admin.GET("/deliveries/performance", h.getDeliveryPerformance)
		admin.GET("/deliveries/time-analysis", h.getDeliveryTimeAnalysis)
		admin.GET("/deliveries/driver-stats", h.getDriverStats)

		// Business insights
		admin.GET("/insights/popular-items", h.getPopularItems)
		admin.GET("/insights/customer-segments", h.getCustomerSegments)
		admin.GET("/insights/market-trends", h.getMarketTrends)
		admin.GET("/insights/recommendations", h.getBusinessRecommendations)

		// Custom reports
		admin.POST("/reports/custom", h.generateCustomReport)
		admin.GET("/reports", h.getReports)
		admin.GET("/reports/:id", h.getReport)
		admin.DELETE("/reports/:id", h.deleteReport)

		// Real-time metrics
		admin.GET("/real-time/orders", h.getRealTimeOrders)
		admin.GET("/real-time/deliveries", h.getRealTimeDeliveries)
		admin.GET("/real-time/revenue", h.getRealTimeRevenue)
	}

	// Merchant analytics
	merchant := router.Group("/merchant/analytics")
	merchant.Use(middleware.AuthMiddleware())
	merchant.Use(middleware.RequireRole(auth.RoleMerchant))
	{
		merchant.GET("/dashboard", h.getMerchantDashboard)
		merchant.GET("/sales", h.getMerchantSales)
		merchant.GET("/orders", h.getMerchantOrders)
		merchant.GET("/popular-items", h.getMerchantPopularItems)
		merchant.GET("/customer-insights", h.getMerchantCustomerInsights)
		merchant.GET("/performance", h.getMerchantPerformance)
	}

	// Driver analytics
	driver := router.Group("/driver/analytics")
	driver.Use(middleware.AuthMiddleware())
	driver.Use(middleware.RequireRole(auth.RoleDriver))
	{
		driver.GET("/dashboard", h.getDriverDashboard)
		driver.GET("/earnings", h.getDriverEarnings)
		driver.GET("/performance", h.getDriverPerformance)
		driver.GET("/delivery-stats", h.getDriverDeliveryStats)
		driver.GET("/ratings", h.getDriverRatings)
	}
}

// Public endpoints

// @Summary Track event
// @Description Track a general analytics event
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body domain.TrackEventRequest true "Event data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/analytics/track-event [post]
func (h *AnalyticsHandler) trackEvent(c *gin.Context) {
	var req domain.TrackEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.analyticsService.TrackEvent(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event tracked successfully"})
}

// @Summary Track order event
// @Description Track an order-related analytics event
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body domain.TrackOrderEventRequest true "Order event data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/analytics/track-order [post]
func (h *AnalyticsHandler) trackOrderEvent(c *gin.Context) {
	var req domain.TrackOrderEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.analyticsService.TrackOrderEvent(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order event tracked successfully"})
}

// @Summary Track delivery event
// @Description Track a delivery-related analytics event
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body domain.TrackDeliveryEventRequest true "Delivery event data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/analytics/track-delivery [post]
func (h *AnalyticsHandler) trackDeliveryEvent(c *gin.Context) {
	var req domain.TrackDeliveryEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.analyticsService.TrackDeliveryEvent(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery event tracked successfully"})
}

// Admin endpoints

// @Summary Get admin dashboard
// @Description Get comprehensive admin dashboard analytics
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.AdminDashboard
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/dashboard [get]
func (h *AnalyticsHandler) getDashboard(c *gin.Context) {
	dashboard, err := h.analyticsService.GetAdminDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// @Summary Get platform statistics
// @Description Get overall platform performance statistics
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.PlatformStats
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/platform-stats [get]
func (h *AnalyticsHandler) getPlatformStats(c *gin.Context) {
	stats, err := h.analyticsService.GetPlatformStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary Get revenue overview
// @Description Get revenue overview analytics
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param period query string false "Time period (day|week|month|year)"
// @Success 200 {object} domain.RevenueOverview
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/revenue/overview [get]
func (h *AnalyticsHandler) getRevenueOverview(c *gin.Context) {
	period := c.DefaultQuery("period", "month")

	overview, err := h.analyticsService.GetRevenueOverview(domain.RevenueRequest{
		Period: period,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// @Summary Get revenue trends
// @Description Get revenue trends over time
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param granularity query string false "Granularity (day|week|month)"
// @Success 200 {object} domain.RevenueTrends
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/revenue/trends [get]
func (h *AnalyticsHandler) getRevenueTrends(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	granularity := c.DefaultQuery("granularity", "day")

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

	req := domain.RevenueTrendsRequest{
		StartDate:   startDate,
		EndDate:     endDate,
		Granularity: granularity,
	}

	trends, err := h.analyticsService.GetRevenueTrends(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trends)
}

// @Summary Get revenue by category
// @Description Get revenue breakdown by product category
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param period query string false "Time period (day|week|month|year)"
// @Success 200 {object} domain.RevenueByCategory
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/revenue/by-category [get]
func (h *AnalyticsHandler) getRevenueByCategory(c *gin.Context) {
	period := c.DefaultQuery("period", "month")

	revenueByCategory, err := h.analyticsService.GetRevenueByCategory(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, revenueByCategory)
}

// @Summary Get revenue by location
// @Description Get revenue breakdown by geographic location
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param period query string false "Time period (day|week|month|year)"
// @Success 200 {object} domain.RevenueByLocation
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/revenue/by-location [get]
func (h *AnalyticsHandler) getRevenueByLocation(c *gin.Context) {
	period := c.DefaultQuery("period", "month")

	revenueByLocation, err := h.analyticsService.GetRevenueByLocation(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, revenueByLocation)
}

// @Summary Get order overview
// @Description Get order analytics overview
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param period query string false "Time period (day|week|month|year)"
// @Success 200 {object} domain.OrderOverview
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/orders/overview [get]
func (h *AnalyticsHandler) getOrderOverview(c *gin.Context) {
	period := c.DefaultQuery("period", "month")

	overview, err := h.analyticsService.GetOrderOverview(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// @Summary Get order trends
// @Description Get order trends over time
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} domain.OrderTrends
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/orders/trends [get]
func (h *AnalyticsHandler) getOrderTrends(c *gin.Context) {
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

	req := domain.OrderTrendsRequest{
		StartDate: startDate,
		EndDate:   endDate,
	}

	trends, err := h.analyticsService.GetOrderTrends(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trends)
}

// @Summary Get orders by status
// @Description Get order count breakdown by status
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param period query string false "Time period (day|week|month|year)"
// @Success 200 {object} domain.OrdersByStatus
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/orders/by-status [get]
func (h *AnalyticsHandler) getOrdersByStatus(c *gin.Context) {
	period := c.DefaultQuery("period", "month")

	ordersByStatus, err := h.analyticsService.GetOrdersByStatus(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ordersByStatus)
}

// @Summary Get peak hours
// @Description Get peak ordering hours analytics
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.PeakHours
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/orders/peak-hours [get]
func (h *AnalyticsHandler) getPeakHours(c *gin.Context) {
	peakHours, err := h.analyticsService.GetPeakHours()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, peakHours)
}

// @Summary Get user overview
// @Description Get user analytics overview
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.UserOverview
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/users/overview [get]
func (h *AnalyticsHandler) getUserOverview(c *gin.Context) {
	overview, err := h.analyticsService.GetUserOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// @Summary Get user activity
// @Description Get user activity patterns
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param period query string false "Time period (day|week|month|year)"
// @Success 200 {object} domain.UserActivity
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/users/activity [get]
func (h *AnalyticsHandler) getUserActivity(c *gin.Context) {
	period := c.DefaultQuery("period", "month")

	activity, err := h.analyticsService.GetUserActivity(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, activity)
}

// @Summary Get user retention
// @Description Get user retention analytics
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.UserRetention
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/users/retention [get]
func (h *AnalyticsHandler) getUserRetention(c *gin.Context) {
	retention, err := h.analyticsService.GetUserRetention()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, retention)
}

// @Summary Get user demographics
// @Description Get user demographic analytics
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.UserDemographics
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/users/demographics [get]
func (h *AnalyticsHandler) getUserDemographics(c *gin.Context) {
	demographics, err := h.analyticsService.GetUserDemographics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, demographics)
}

// @Summary Get delivery overview
// @Description Get delivery analytics overview
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.DeliveryOverview
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/deliveries/overview [get]
func (h *AnalyticsHandler) getDeliveryOverview(c *gin.Context) {
	overview, err := h.analyticsService.GetDeliveryOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// @Summary Get delivery performance
// @Description Get delivery performance metrics
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param period query string false "Time period (day|week|month|year)"
// @Success 200 {object} domain.DeliveryPerformance
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/deliveries/performance [get]
func (h *AnalyticsHandler) getDeliveryPerformance(c *gin.Context) {
	period := c.DefaultQuery("period", "month")

	performance, err := h.analyticsService.GetDeliveryPerformance(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, performance)
}

// @Summary Get delivery time analysis
// @Description Get delivery time analysis
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.DeliveryTimeAnalysis
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/deliveries/time-analysis [get]
func (h *AnalyticsHandler) getDeliveryTimeAnalysis(c *gin.Context) {
	analysis, err := h.analyticsService.GetDeliveryTimeAnalysis()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// @Summary Get driver statistics
// @Description Get driver performance statistics
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.DriverStats
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/deliveries/driver-stats [get]
func (h *AnalyticsHandler) getDriverStats(c *gin.Context) {
	stats, err := h.analyticsService.GetDriverStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary Get popular items
// @Description Get most popular items insights
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results"
// @Param period query string false "Time period"
// @Success 200 {object} domain.PopularItems
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/insights/popular-items [get]
func (h *AnalyticsHandler) getPopularItems(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	period := c.DefaultQuery("period", "month")

	limit, _ := strconv.Atoi(limitStr)

	req := domain.PopularItemsRequest{
		Limit:  limit,
		Period: period,
	}

	items, err := h.analyticsService.GetPopularItems(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

// @Summary Get customer segments
// @Description Get customer segmentation insights
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.CustomerSegments
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/insights/customer-segments [get]
func (h *AnalyticsHandler) getCustomerSegments(c *gin.Context) {
	segments, err := h.analyticsService.GetCustomerSegments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, segments)
}

// @Summary Get market trends
// @Description Get market trends insights
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.MarketTrends
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/insights/market-trends [get]
func (h *AnalyticsHandler) getMarketTrends(c *gin.Context) {
	trends, err := h.analyticsService.GetMarketTrends()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, trends)
}

// @Summary Get business recommendations
// @Description Get AI-powered business recommendations
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.BusinessRecommendations
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/insights/recommendations [get]
func (h *AnalyticsHandler) getBusinessRecommendations(c *gin.Context) {
	recommendations, err := h.analyticsService.GetBusinessRecommendations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, recommendations)
}

// @Summary Generate custom report
// @Description Generate a custom analytics report
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CustomReportRequest true "Report configuration"
// @Success 201 {object} domain.AnalyticsReport
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/reports/custom [post]
func (h *AnalyticsHandler) generateCustomReport(c *gin.Context) {
	var req domain.CustomReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID, _ := c.Get("user_id")
	req.CreatedBy = adminID.(string)

	report, err := h.analyticsService.GenerateCustomReport(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, report)
}

// @Summary Get reports
// @Description Get list of generated reports
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.AnalyticsReport
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/reports [get]
func (h *AnalyticsHandler) getReports(c *gin.Context) {
	adminID, _ := c.Get("user_id")

	reports, err := h.analyticsService.GetReports(adminID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reports)
}

// @Summary Get report
// @Description Get a specific report by ID
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Report ID"
// @Success 200 {object} domain.AnalyticsReport
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/reports/{id} [get]
func (h *AnalyticsHandler) getReport(c *gin.Context) {
	reportID := c.Param("id")
	adminID, _ := c.Get("user_id")

	report, err := h.analyticsService.GetReport(reportID, adminID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// @Summary Delete report
// @Description Delete a report
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Report ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/reports/{id} [delete]
func (h *AnalyticsHandler) deleteReport(c *gin.Context) {
	reportID := c.Param("id")
	adminID, _ := c.Get("user_id")

	err := h.analyticsService.DeleteReport(reportID, adminID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Report deleted successfully"})
}

// Real-time endpoints

// @Summary Get real-time orders
// @Description Get real-time order metrics
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.RealTimeOrders
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/real-time/orders [get]
func (h *AnalyticsHandler) getRealTimeOrders(c *gin.Context) {
	realTimeOrders, err := h.analyticsService.GetRealTimeOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, realTimeOrders)
}

// @Summary Get real-time deliveries
// @Description Get real-time delivery metrics
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.RealTimeDeliveries
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/real-time/deliveries [get]
func (h *AnalyticsHandler) getRealTimeDeliveries(c *gin.Context) {
	realTimeDeliveries, err := h.analyticsService.GetRealTimeDeliveries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, realTimeDeliveries)
}

// @Summary Get real-time revenue
// @Description Get real-time revenue metrics
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.RealTimeRevenue
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/analytics/real-time/revenue [get]
func (h *AnalyticsHandler) getRealTimeRevenue(c *gin.Context) {
	realTimeRevenue, err := h.analyticsService.GetRealTimeRevenue()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, realTimeRevenue)
}

// Merchant endpoints

// @Summary Get merchant dashboard
// @Description Get merchant analytics dashboard
// @Tags merchant
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.MerchantDashboard
// @Failure 500 {object} map[string]string
// @Router /api/v1/merchant/analytics/dashboard [get]
func (h *AnalyticsHandler) getMerchantDashboard(c *gin.Context) {
	merchantID, _ := c.Get("user_id")

	dashboard, err := h.analyticsService.GetMerchantDashboard(merchantID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// @Summary Get merchant sales
// @Description Get merchant sales analytics
// @Tags merchant
// @Produce json
// @Security BearerAuth
// @Param period query string false "Time period"
// @Success 200 {object} domain.MerchantSales
// @Failure 500 {object} map[string]string
// @Router /api/v1/merchant/analytics/sales [get]
func (h *AnalyticsHandler) getMerchantSales(c *gin.Context) {
	merchantID, _ := c.Get("user_id")
	period := c.DefaultQuery("period", "month")

	req := domain.MerchantAnalyticsRequest{
		MerchantID: merchantID.(string),
		Period:     period,
	}

	sales, err := h.analyticsService.GetMerchantSales(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sales)
}

// @Summary Get merchant orders
// @Description Get merchant order analytics
// @Tags merchant
// @Produce json
// @Security BearerAuth
// @Param period query string false "Time period"
// @Success 200 {object} domain.MerchantOrders
// @Failure 500 {object} map[string]string
// @Router /api/v1/merchant/analytics/orders [get]
func (h *AnalyticsHandler) getMerchantOrders(c *gin.Context) {
	merchantID, _ := c.Get("user_id")
	period := c.DefaultQuery("period", "month")

	req := domain.MerchantAnalyticsRequest{
		MerchantID: merchantID.(string),
		Period:     period,
	}

	orders, err := h.analyticsService.GetMerchantOrders(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// @Summary Get merchant popular items
// @Description Get merchant's most popular items
// @Tags merchant
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results"
// @Success 200 {object} domain.MerchantPopularItems
// @Failure 500 {object} map[string]string
// @Router /api/v1/merchant/analytics/popular-items [get]
func (h *AnalyticsHandler) getMerchantPopularItems(c *gin.Context) {
	merchantID, _ := c.Get("user_id")
	limitStr := c.DefaultQuery("limit", "10")

	limit, _ := strconv.Atoi(limitStr)

	items, err := h.analyticsService.GetMerchantPopularItems(merchantID.(string), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

// @Summary Get merchant customer insights
// @Description Get merchant customer analytics insights
// @Tags merchant
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.MerchantCustomerInsights
// @Failure 500 {object} map[string]string
// @Router /api/v1/merchant/analytics/customer-insights [get]
func (h *AnalyticsHandler) getMerchantCustomerInsights(c *gin.Context) {
	merchantID, _ := c.Get("user_id")

	insights, err := h.analyticsService.GetMerchantCustomerInsights(merchantID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, insights)
}

// @Summary Get merchant performance
// @Description Get merchant performance metrics
// @Tags merchant
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.MerchantPerformance
// @Failure 500 {object} map[string]string
// @Router /api/v1/merchant/analytics/performance [get]
func (h *AnalyticsHandler) getMerchantPerformance(c *gin.Context) {
	merchantID, _ := c.Get("user_id")

	performance, err := h.analyticsService.GetMerchantPerformance(merchantID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, performance)
}

// Driver endpoints

// @Summary Get driver dashboard
// @Description Get driver analytics dashboard
// @Tags driver
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.DriverDashboard
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/analytics/dashboard [get]
func (h *AnalyticsHandler) getDriverDashboard(c *gin.Context) {
	driverID, _ := c.Get("user_id")

	dashboard, err := h.analyticsService.GetDriverDashboard(driverID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// @Summary Get driver earnings
// @Description Get driver earnings analytics
// @Tags driver
// @Produce json
// @Security BearerAuth
// @Param period query string false "Time period"
// @Success 200 {object} domain.DriverEarnings
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/analytics/earnings [get]
func (h *AnalyticsHandler) getDriverEarnings(c *gin.Context) {
	driverID, _ := c.Get("user_id")
	period := c.DefaultQuery("period", "month")

	req := domain.DriverAnalyticsRequest{
		DriverID: driverID.(string),
		Period:   period,
	}

	earnings, err := h.analyticsService.GetDriverEarnings(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, earnings)
}

// @Summary Get driver performance
// @Description Get driver performance analytics
// @Tags driver
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.DriverPerformanceAnalytics
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/analytics/performance [get]
func (h *AnalyticsHandler) getDriverPerformance(c *gin.Context) {
	driverID, _ := c.Get("user_id")

	performance, err := h.analyticsService.GetDriverPerformanceAnalytics(driverID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, performance)
}

// @Summary Get driver delivery stats
// @Description Get driver delivery statistics
// @Tags driver
// @Produce json
// @Security BearerAuth
// @Param period query string false "Time period"
// @Success 200 {object} domain.DriverDeliveryStats
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/analytics/delivery-stats [get]
func (h *AnalyticsHandler) getDriverDeliveryStats(c *gin.Context) {
	driverID, _ := c.Get("user_id")
	period := c.DefaultQuery("period", "month")

	req := domain.DriverAnalyticsRequest{
		DriverID: driverID.(string),
		Period:   period,
	}

	stats, err := h.analyticsService.GetDriverDeliveryStats(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary Get driver ratings
// @Description Get driver ratings analytics
// @Tags driver
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.DriverRatings
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/analytics/ratings [get]
func (h *AnalyticsHandler) getDriverRatings(c *gin.Context) {
	driverID, _ := c.Get("user_id")

	ratings, err := h.analyticsService.GetDriverRatings(driverID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ratings)
}

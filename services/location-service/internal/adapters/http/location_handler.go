package http

import (
	"net/http"
	"strconv"
	"time"

	"glovo-backend/services/location-service/internal/domain"
	"glovo-backend/shared/auth"
	"glovo-backend/shared/middleware"

	"github.com/gin-gonic/gin"
)

type LocationHandler struct {
	locationService domain.LocationService
}

func NewLocationHandler(locationService domain.LocationService) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
	}
}

func (h *LocationHandler) SetupRoutes(router *gin.RouterGroup) {
	// Driver location routes (authenticated drivers)
	drivers := router.Group("/drivers")
	drivers.Use(middleware.AuthMiddleware())
	drivers.Use(middleware.RequireRole(auth.RoleDriver))
	{
		// Update driver location
		drivers.PUT("/location", h.updateDriverLocation)

		// Get driver's own location
		drivers.GET("/location", h.getDriverLocation)

		// Get driver's location history
		drivers.GET("/location/history", h.getDriverLocationHistory)
	}

	// Delivery route management (authenticated drivers)
	routes := router.Group("/routes")
	routes.Use(middleware.AuthMiddleware())
	routes.Use(middleware.RequireRole(auth.RoleDriver))
	{
		// Create new delivery route
		routes.POST("/", h.createDeliveryRoute)

		// Get current route
		routes.GET("/current", h.getCurrentRoute)

		// Update route progress
		routes.PUT("/progress", h.updateRouteProgress)

		// Complete route
		routes.PUT("/complete", h.completeRoute)

		// Get route history
		routes.GET("/history", h.getRouteHistory)
	}

	// Public location search endpoints
	search := router.Group("/search")
	{
		// Find nearby drivers (for order assignment)
		search.GET("/drivers/nearby", h.findNearbyDrivers)

		// Calculate delivery estimate
		search.POST("/estimate", h.calculateDeliveryEstimate)
	}

	// Admin geofence management
	geofences := router.Group("/geofences")
	geofences.Use(middleware.AuthMiddleware())
	geofences.Use(middleware.RequireRole(auth.RoleAdmin))
	{
		geofences.POST("/", h.createGeofence)
		geofences.GET("/", h.listGeofences)
	}

	// Analytics endpoints (admin only)
	analytics := router.Group("/analytics")
	analytics.Use(middleware.AuthMiddleware())
	analytics.Use(middleware.RequireRole(auth.RoleAdmin))
	{
		analytics.GET("/driver-activity", h.getDriverActivity)
		analytics.GET("/delivery-patterns", h.getDeliveryPatterns)
		analytics.GET("/coverage-zones", h.getCoverageZones)
	}
}

// @Summary Update driver location
// @Description Update the current location of a driver
// @Tags drivers
// @Accept json
// @Produce json
// @Param request body domain.UpdateLocationRequest true "Location update request"
// @Success 200 {object} domain.DriverLocation
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /drivers/location [put]
func (h *LocationHandler) updateDriverLocation(c *gin.Context) {
	userID := c.GetString("user_id")

	var req domain.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	location, err := h.locationService.UpdateDriverLocation(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, location)
}

// @Summary Get driver location
// @Description Get the current location of a driver
// @Tags drivers
// @Produce json
// @Success 200 {object} domain.DriverLocation
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /drivers/location [get]
func (h *LocationHandler) getDriverLocation(c *gin.Context) {
	userID := c.GetString("user_id")

	location, err := h.locationService.GetDriverLocation(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Location not found"})
		return
	}

	c.JSON(http.StatusOK, location)
}

// @Summary Get driver location history
// @Description Get the location history of a driver
// @Tags drivers
// @Produce json
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.LocationHistory
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /drivers/location/history [get]
func (h *LocationHandler) getDriverLocationHistory(c *gin.Context) {
	userID := c.GetString("user_id")

	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time and end_time are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, use RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, use RFC3339"})
		return
	}

	req := domain.LocationHistoryRequest{
		DriverID:  userID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	if orderID := c.Query("order_id"); orderID != "" {
		req.OrderID = &orderID
	}

	history, err := h.locationService.GetLocationHistory(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// @Summary Create delivery route
// @Description Create a new delivery route for a driver
// @Tags routes
// @Accept json
// @Produce json
// @Param request body domain.RouteRequest true "Route creation request"
// @Success 201 {object} domain.DeliveryRoute
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /routes [post]
func (h *LocationHandler) createDeliveryRoute(c *gin.Context) {
	userID := c.GetString("user_id")

	var req domain.RouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.DriverID = userID
	route, err := h.locationService.CreateDeliveryRoute(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, route)
}

// @Summary Get current route
// @Description Get the current active routes (returns all active routes)
// @Tags routes
// @Produce json
// @Success 200 {array} domain.DeliveryRoute
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /routes/current [get]
func (h *LocationHandler) getCurrentRoute(c *gin.Context) {
	routes, err := h.locationService.GetActiveRoutes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, routes)
}

// @Summary Update route progress
// @Description Update the progress of a delivery route
// @Tags routes
// @Accept json
// @Produce json
// @Param request body struct{OrderID string `json:"order_id"`; CurrentLocation domain.GeoPoint `json:"current_location"`} true "Route progress update"
// @Success 200 {object} domain.DeliveryRoute
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /routes/progress [put]
func (h *LocationHandler) updateRouteProgress(c *gin.Context) {
	var req struct {
		OrderID         string          `json:"order_id" binding:"required"`
		CurrentLocation domain.GeoPoint `json:"current_location" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	route, err := h.locationService.UpdateRouteProgress(req.OrderID, req.CurrentLocation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, route)
}

// @Summary Complete route
// @Description Mark a delivery route as completed
// @Tags routes
// @Accept json
// @Produce json
// @Param request body struct{OrderID string `json:"order_id"`} true "Route completion data"
// @Success 200 {object} domain.DeliveryRoute
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /routes/complete [put]
func (h *LocationHandler) completeRoute(c *gin.Context) {
	var req struct {
		OrderID string `json:"order_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	route, err := h.locationService.CompleteRoute(req.OrderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, route)
}

// @Summary Get route history
// @Description Get all active routes (simplified endpoint)
// @Tags routes
// @Produce json
// @Success 200 {array} domain.DeliveryRoute
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /routes/history [get]
func (h *LocationHandler) getRouteHistory(c *gin.Context) {
	routes, err := h.locationService.GetActiveRoutes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, routes)
}

// @Summary Find nearby drivers
// @Description Find drivers near a specific location
// @Tags search
// @Produce json
// @Param latitude query float64 true "Latitude"
// @Param longitude query float64 true "Longitude"
// @Param radius query float64 false "Search radius in kilometers" default(5)
// @Param limit query int false "Limit" default(10)
// @Success 200 {array} domain.NearbyDriver
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /search/drivers/nearby [get]
func (h *LocationHandler) findNearbyDrivers(c *gin.Context) {
	latStr := c.Query("latitude")
	lonStr := c.Query("longitude")
	radiusStr := c.DefaultQuery("radius", "5")
	limitStr := c.DefaultQuery("limit", "10")

	if latStr == "" || lonStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "latitude and longitude are required"})
		return
	}

	latitude, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid latitude"})
		return
	}

	longitude, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid longitude"})
		return
	}

	radius, _ := strconv.ParseFloat(radiusStr, 64)
	limit, _ := strconv.Atoi(limitStr)

	req := domain.NearbyDriversRequest{
		Latitude:  latitude,
		Longitude: longitude,
		Radius:    radius,
		Limit:     limit,
	}

	drivers, err := h.locationService.GetNearbyDrivers(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, drivers)
}

// @Summary Calculate delivery estimate
// @Description Calculate estimated delivery time and route (simplified mock)
// @Tags search
// @Accept json
// @Produce json
// @Param request body struct{OriginLat float64 `json:"origin_lat"`; OriginLng float64 `json:"origin_lng"`; DestLat float64 `json:"dest_lat"`; DestLng float64 `json:"dest_lng"`} true "Estimation request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /search/estimate [post]
func (h *LocationHandler) calculateDeliveryEstimate(c *gin.Context) {
	var req struct {
		OriginLat float64 `json:"origin_lat" binding:"required"`
		OriginLng float64 `json:"origin_lng" binding:"required"`
		DestLat   float64 `json:"dest_lat" binding:"required"`
		DestLng   float64 `json:"dest_lng" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock delivery estimate
	estimate := map[string]interface{}{
		"estimated_duration_minutes": 25,
		"estimated_distance_km":      8.5,
		"estimated_arrival_time":     time.Now().Add(25 * time.Minute),
	}

	c.JSON(http.StatusOK, estimate)
}

// @Summary Create geofence
// @Description Create a new geofence area
// @Tags geofences
// @Accept json
// @Produce json
// @Param request body domain.Geofence true "Geofence creation request"
// @Success 201 {object} domain.Geofence
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /geofences [post]
func (h *LocationHandler) createGeofence(c *gin.Context) {
	var geofence domain.Geofence
	if err := c.ShouldBindJSON(&geofence); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.locationService.CreateGeofence(&geofence)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// @Summary List geofences
// @Description Get all geofences
// @Tags geofences
// @Produce json
// @Success 200 {array} domain.Geofence
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /geofences [get]
func (h *LocationHandler) listGeofences(c *gin.Context) {
	geofences, err := h.locationService.GetGeofences()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, geofences)
}

// Analytics endpoints
func (h *LocationHandler) getDriverActivity(c *gin.Context) {
	// Implementation for driver activity analytics
	c.JSON(http.StatusOK, gin.H{"message": "Driver activity analytics endpoint"})
}

func (h *LocationHandler) getDeliveryPatterns(c *gin.Context) {
	// Implementation for delivery patterns analytics
	c.JSON(http.StatusOK, gin.H{"message": "Delivery patterns analytics endpoint"})
}

func (h *LocationHandler) getCoverageZones(c *gin.Context) {
	// Implementation for coverage zones analytics
	c.JSON(http.StatusOK, gin.H{"message": "Coverage zones analytics endpoint"})
}

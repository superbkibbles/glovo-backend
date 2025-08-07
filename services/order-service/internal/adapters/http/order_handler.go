package http

import (
	"net/http"
	"strconv"

	"glovo-backend/services/order-service/internal/domain"
	"glovo-backend/shared/auth"
	"glovo-backend/shared/middleware"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService domain.OrderService
}

func NewOrderHandler(orderService domain.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) SetupRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		// Customer routes
		customer := v1.Group("/orders")
		customer.Use(middleware.AuthMiddleware())
		customer.Use(middleware.RequireRoles([]auth.UserRole{auth.RoleCustomer, auth.RoleAdmin}))
		{
			customer.POST("", h.CreateOrder)
			customer.GET("", h.GetOrderHistory)
			customer.GET("/:id", h.GetOrder)
			customer.PUT("/:id/cancel", h.CancelOrder)
		}

		// Merchant routes
		merchant := v1.Group("/merchant/orders")
		merchant.Use(middleware.AuthMiddleware())
		merchant.Use(middleware.RequireRoles([]auth.UserRole{auth.RoleMerchant, auth.RoleAdmin}))
		{
			merchant.GET("", h.GetMerchantOrders)
			merchant.GET("/:id", h.GetOrder)
			merchant.PUT("/:id/status", h.UpdateOrderStatus)
		}

		// Driver routes
		driver := v1.Group("/driver/orders")
		driver.Use(middleware.AuthMiddleware())
		driver.Use(middleware.RequireRoles([]auth.UserRole{auth.RoleDriver, auth.RoleAdmin}))
		{
			driver.GET("", h.GetDriverOrders)
			driver.GET("/:id", h.GetOrder)
			driver.PUT("/:id/status", h.UpdateOrderStatus)
		}

		// Admin routes
		admin := v1.Group("/admin/orders")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.RequireRole(auth.RoleAdmin))
		{
			admin.GET("", h.GetAllOrders)
			admin.GET("/active", h.GetActiveOrders)
			admin.GET("/:id", h.GetOrder)
			admin.PUT("/:id/status", h.UpdateOrderStatus)
		}
	}
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order for the authenticated customer
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateOrderRequest true "Order details"
// @Success 201 {object} domain.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req domain.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.orderService.CreateOrder(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetOrder godoc
// @Summary Get order by ID
// @Description Get order details by order ID
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} domain.OrderResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	userID := c.GetString("user_id")
	role := auth.UserRole(c.GetString("role"))

	response, err := h.orderService.GetOrder(orderID, userID, role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetOrderHistory godoc
// @Summary Get customer order history
// @Description Get paginated order history for the authenticated customer
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.Order
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders [get]
func (h *OrderHandler) GetOrderHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	role := auth.UserRole(c.GetString("role"))

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	orders, err := h.orderService.GetOrderHistory(userID, role, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Update the status of an order (merchant/driver/admin only)
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body domain.UpdateOrderStatusRequest true "Status update"
// @Success 200 {object} domain.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/orders/{id}/status [put]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	userID := c.GetString("user_id")
	role := auth.UserRole(c.GetString("role"))

	var req domain.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.orderService.UpdateOrderStatus(orderID, req, userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CancelOrder godoc
// @Summary Cancel an order
// @Description Cancel an order (customer/admin only)
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body map[string]string true "Cancellation reason"
// @Success 200 {object} domain.OrderResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/orders/{id}/cancel [put]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	userID := c.GetString("user_id")
	role := auth.UserRole(c.GetString("role"))

	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reason := req["reason"]
	if reason == "" {
		reason = "Cancelled by customer"
	}

	response, err := h.orderService.CancelOrder(orderID, userID, role, reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetMerchantOrders godoc
// @Summary Get merchant orders
// @Description Get orders for the authenticated merchant
// @Tags Merchant
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.Order
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/merchant/orders [get]
func (h *OrderHandler) GetMerchantOrders(c *gin.Context) {
	merchantID := c.GetString("user_id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	orders, err := h.orderService.GetOrdersForMerchant(merchantID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetDriverOrders godoc
// @Summary Get driver orders
// @Description Get orders assigned to the authenticated driver
// @Tags Driver
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.Order
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/driver/orders [get]
func (h *OrderHandler) GetDriverOrders(c *gin.Context) {
	driverID := c.GetString("user_id")

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	orders, err := h.orderService.GetOrdersForDriver(driverID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetAllOrders godoc
// @Summary Get all orders (Admin only)
// @Description Get all orders in the system with pagination
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.Order
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/orders [get]
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	userID := c.GetString("user_id")
	role := auth.UserRole(c.GetString("role"))

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	orders, err := h.orderService.GetOrderHistory(userID, role, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetActiveOrders godoc
// @Summary Get active orders (Admin only)
// @Description Get all active orders in the system
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} domain.Order
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/orders/active [get]
func (h *OrderHandler) GetActiveOrders(c *gin.Context) {
	orders, err := h.orderService.GetActiveOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

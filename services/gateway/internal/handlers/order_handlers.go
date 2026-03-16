package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mendmzury/food-delivery/pkg/config"
	commonpb "github.com/mendmzury/food-delivery/proto/common"
	orderpb "github.com/mendmzury/food-delivery/proto/order"
	"github.com/mendmzury/food-delivery/services/gateway/internal/grpc"
)

// ListOrders godoc
// @Summary List orders
// @Tags orders
// @Security BearerAuth
// @Success 200 {object} object
// @Router /orders [get]
func ListOrders(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, pageSize := getPaginationParams(c)
		customerID := c.Query("customer_id")
		restaurantID := c.Query("restaurant_id")
		driverID := c.Query("driver_id")
		statusStr := c.Query("status")

		status := orderpb.OrderStatus_ORDER_STATUS_PENDING
		switch statusStr {
		case "accepted":
			status = orderpb.OrderStatus_ORDER_STATUS_ACCEPTED
		case "preparing":
			status = orderpb.OrderStatus_ORDER_STATUS_PREPARING
		case "ready":
			status = orderpb.OrderStatus_ORDER_STATUS_READY
		case "assigned":
			status = orderpb.OrderStatus_ORDER_STATUS_ASSIGNED
		case "picked_up":
			status = orderpb.OrderStatus_ORDER_STATUS_PICKED_UP
		case "delivered":
			status = orderpb.OrderStatus_ORDER_STATUS_DELIVERED
		case "cancelled":
			status = orderpb.OrderStatus_ORDER_STATUS_CANCELLED
		default:
			status = 0
		}

		ctx := getAuthContext(c)
		resp, err := clients.Order.ListOrders(ctx, &orderpb.ListOrdersRequest{
			Pagination:   &commonpb.PaginationRequest{Page: page, PageSize: pageSize},
			CustomerId:   customerID,
			RestaurantId: restaurantID,
			DriverId:     driverID,
			Status:       status,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		paginatedResponse(c, resp.Orders, resp.Pagination.Total, page, pageSize)
	}
}

// GetOrder godoc
// @Summary Get order
// @Tags orders
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} object
// @Router /orders/{id} [get]
func GetOrder(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx := getAuthContext(c)
		order, err := clients.Order.GetOrder(ctx, &orderpb.GetOrderRequest{Id: id})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, order)
	}
}

// CreateOrder godoc
// @Summary Create order
// @Tags orders
// @Security BearerAuth
// @Success 201 {object} object
// @Router /orders [post]
func CreateOrder(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			CustomerID      string `json:"customer_id" binding:"required"`
			RestaurantID    string `json:"restaurant_id" binding:"required"`
			DeliveryAddress string `json:"delivery_address" binding:"required"`
			DeliveryLat     float64 `json:"delivery_lat"`
			DeliveryLng     float64 `json:"delivery_lng"`
			Notes           string  `json:"notes"`
			Items           []struct {
				MenuItemID string `json:"menu_item_id" binding:"required"`
				Quantity   int32  `json:"quantity" binding:"required"`
				Options    string `json:"options"`
			} `json:"items" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		items := make([]*orderpb.OrderItemInput, len(req.Items))
		for i, it := range req.Items {
			items[i] = &orderpb.OrderItemInput{
				MenuItemId: it.MenuItemID,
				Quantity:   it.Quantity,
				Options:   it.Options,
			}
		}
		ctx := getAuthContext(c)
		order, err := clients.Order.CreateOrder(ctx, &orderpb.CreateOrderRequest{
			CustomerId:      req.CustomerID,
			RestaurantId:    req.RestaurantID,
			DeliveryAddress: req.DeliveryAddress,
			DeliveryLat:     req.DeliveryLat,
			DeliveryLng:     req.DeliveryLng,
			Notes:           req.Notes,
			Items:           items,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		createdResponse(c, order)
	}
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Tags orders
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} object
// @Router /orders/{id}/status [put]
func UpdateOrderStatus(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Status string `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		status := orderpb.OrderStatus_ORDER_STATUS_PENDING
		switch req.Status {
		case "accepted":
			status = orderpb.OrderStatus_ORDER_STATUS_ACCEPTED
		case "preparing":
			status = orderpb.OrderStatus_ORDER_STATUS_PREPARING
		case "ready":
			status = orderpb.OrderStatus_ORDER_STATUS_READY
		case "assigned":
			status = orderpb.OrderStatus_ORDER_STATUS_ASSIGNED
		case "picked_up":
			status = orderpb.OrderStatus_ORDER_STATUS_PICKED_UP
		case "delivered":
			status = orderpb.OrderStatus_ORDER_STATUS_DELIVERED
		case "cancelled":
			status = orderpb.OrderStatus_ORDER_STATUS_CANCELLED
		default:
			errorResponse(c, http.StatusBadRequest, "invalid status")
			return
		}
		ctx := getAuthContext(c)
		order, err := clients.Order.UpdateOrderStatus(ctx, &orderpb.UpdateOrderStatusRequest{
			Id:     id,
			Status: status,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, order)
	}
}

// CancelOrder godoc
// @Summary Cancel order
// @Tags orders
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} object
// @Router /orders/{id}/cancel [post]
func CancelOrder(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx := getAuthContext(c)
		order, err := clients.Order.CancelOrder(ctx, &orderpb.CancelOrderRequest{Id: id})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, order)
	}
}

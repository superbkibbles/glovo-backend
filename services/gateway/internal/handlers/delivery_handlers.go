package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mendmzury/food-delivery/pkg/config"
	commonpb "github.com/mendmzury/food-delivery/proto/common"
	deliverypb "github.com/mendmzury/food-delivery/proto/delivery"
	"github.com/mendmzury/food-delivery/services/gateway/internal/grpc"
)

// ListAssignments godoc
// @Summary List delivery assignments
// @Tags delivery
// @Security BearerAuth
// @Success 200 {object} object
// @Router /delivery/assignments [get]
func ListAssignments(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, pageSize := getPaginationParams(c)
		driverID := c.Query("driver_id")
		statusStr := c.Query("status")

		status := deliverypb.AssignmentStatus_ASSIGNMENT_STATUS_ASSIGNED
		switch statusStr {
		case "picked_up":
			status = deliverypb.AssignmentStatus_ASSIGNMENT_STATUS_PICKED_UP
		case "delivered":
			status = deliverypb.AssignmentStatus_ASSIGNMENT_STATUS_DELIVERED
		default:
			status = 0
		}

		ctx := getAuthContext(c)
		resp, err := clients.Delivery.ListAssignments(ctx, &deliverypb.ListAssignmentsRequest{
			Pagination: &commonpb.PaginationRequest{Page: page, PageSize: pageSize},
			DriverId:   driverID,
			Status:     status,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		paginatedResponse(c, resp.Assignments, resp.Pagination.Total, page, pageSize)
	}
}

// GetAssignment godoc
// @Summary Get assignment by order_id
// @Tags delivery
// @Security BearerAuth
// @Success 200 {object} object
// @Router /delivery/assignments/by [get]
func GetAssignment(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Query("id")
		orderID := c.Query("order_id")
		if id == "" && orderID == "" {
			errorResponse(c, http.StatusBadRequest, "id or order_id required")
			return
		}
		ctx := getAuthContext(c)
		a, err := clients.Delivery.GetAssignment(ctx, &deliverypb.GetAssignmentRequest{
			Id:      id,
			OrderId: orderID,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, a)
	}
}

// AssignDriver godoc
// @Summary Assign driver to order
// @Tags delivery
// @Security BearerAuth
// @Success 201 {object} object
// @Router /delivery/assign [post]
func AssignDriver(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OrderID  string `json:"order_id" binding:"required"`
			DriverID string `json:"driver_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		ctx := getAuthContext(c)
		a, err := clients.Delivery.AssignDriver(ctx, &deliverypb.AssignDriverRequest{
			OrderId:  req.OrderID,
			DriverId: req.DriverID,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		createdResponse(c, a)
	}
}

// UpdateDriverLocation godoc
// @Summary Update driver location
// @Tags delivery
// @Security BearerAuth
// @Success 200 {object} object
// @Router /delivery/location [post]
func UpdateDriverLocation(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			DriverID string  `json:"driver_id" binding:"required"`
			Lat      float64 `json:"lat" binding:"required"`
			Lng      float64 `json:"lng" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		ctx := getAuthContext(c)
		_, err := clients.Delivery.UpdateLocation(ctx, &deliverypb.UpdateLocationRequest{
			DriverId: req.DriverID,
			Lat:      req.Lat,
			Lng:      req.Lng,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, gin.H{"message": "updated"})
	}
}

// CompleteDelivery godoc
// @Summary Complete delivery
// @Tags delivery
// @Security BearerAuth
// @Param id path string true "Assignment ID"
// @Success 200 {object} object
// @Router /delivery/assignments/{id}/complete [post]
func CompleteDelivery(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		assignmentID := c.Param("id")
		ctx := getAuthContext(c)
		a, err := clients.Delivery.CompleteDelivery(ctx, &deliverypb.CompleteDeliveryRequest{
			AssignmentId: assignmentID,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, a)
	}
}

// GetTrackingHistory godoc
// @Summary Get tracking history
// @Tags delivery
// @Security BearerAuth
// @Param id path string true "Assignment ID"
// @Success 200 {object} object
// @Router /delivery/assignments/{id}/tracking-history [get]
func GetTrackingHistory(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		assignmentID := c.Param("id")
		ctx := getAuthContext(c)
		resp, err := clients.Delivery.GetTrackingHistory(ctx, &deliverypb.GetTrackingHistoryRequest{
			AssignmentId: assignmentID,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, resp.Points)
	}
}

// ListDriverLocations godoc
// @Summary List driver locations
// @Tags delivery
// @Security BearerAuth
// @Success 200 {object} object
// @Router /delivery/driver-locations [get]
func ListDriverLocations(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := getAuthContext(c)
		resp, err := clients.Delivery.ListDriverLocations(ctx, &deliverypb.ListDriverLocationsRequest{})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, resp.Locations)
	}
}

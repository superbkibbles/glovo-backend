package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mendmzury/food-delivery/pkg/config"
	settingspb "github.com/mendmzury/food-delivery/proto/settings"
	"github.com/mendmzury/food-delivery/services/gateway/internal/grpc"
)

// GetCommission godoc
// @Summary Get commission
// @Tags settings
// @Security BearerAuth
// @Success 200 {object} object
// @Router /settings/commission [get]
func GetCommission(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := getAuthContext(c)
		resp, err := clients.Settings.GetCommission(ctx, &settingspb.GetCommissionRequest{})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, gin.H{"commission_percent": resp.CommissionPercent})
	}
}

// UpdateCommission godoc
// @Summary Update commission
// @Tags settings
// @Security BearerAuth
// @Success 200 {object} object
// @Router /settings/commission [put]
func UpdateCommission(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			CommissionPercent float64 `json:"commission_percent" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		ctx := getAuthContext(c)
		resp, err := clients.Settings.UpdateCommission(ctx, &settingspb.UpdateCommissionRequest{
			CommissionPercent: req.CommissionPercent,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, gin.H{"commission_percent": resp.CommissionPercent})
	}
}

// ListOperatingAreas godoc
// @Summary List operating areas
// @Tags settings
// @Security BearerAuth
// @Success 200 {object} object
// @Router /settings/operating-areas [get]
func ListOperatingAreas(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := getAuthContext(c)
		resp, err := clients.Settings.ListOperatingAreas(ctx, &settingspb.ListOperatingAreasRequest{})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		areas := make([]gin.H, len(resp.Areas))
		for i, a := range resp.Areas {
			areas[i] = areaToJSON(a)
		}
		successResponse(c, areas)
	}
}

// CreateOperatingArea godoc
// @Summary Create operating area
// @Tags settings
// @Security BearerAuth
// @Success 201 {object} object
// @Router /settings/operating-areas [post]
func CreateOperatingArea(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string  `json:"name" binding:"required"`
			Coordinates []coord `json:"coordinates" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		coords := make([]*settingspb.Coordinate, len(req.Coordinates))
		for i, c := range req.Coordinates {
			coords[i] = &settingspb.Coordinate{Lng: c.Lng, Lat: c.Lat}
		}
		ctx := getAuthContext(c)
		resp, err := clients.Settings.CreateOperatingArea(ctx, &settingspb.CreateOperatingAreaRequest{
			Name:        req.Name,
			Coordinates: coords,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		createdResponse(c, areaToJSON(resp))
	}
}

// UpdateOperatingArea godoc
// @Summary Update operating area
// @Tags settings
// @Security BearerAuth
// @Param id path string true "Area ID"
// @Success 200 {object} object
// @Router /settings/operating-areas/{id} [put]
func UpdateOperatingArea(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Name        string  `json:"name"`
			Coordinates []coord `json:"coordinates"`
			Active      bool    `json:"active"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		var coords []*settingspb.Coordinate
		if len(req.Coordinates) >= 3 {
			coords = make([]*settingspb.Coordinate, len(req.Coordinates))
			for i, c := range req.Coordinates {
				coords[i] = &settingspb.Coordinate{Lng: c.Lng, Lat: c.Lat}
			}
		}
		ctx := getAuthContext(c)
		resp, err := clients.Settings.UpdateOperatingArea(ctx, &settingspb.UpdateOperatingAreaRequest{
			Id:          id,
			Name:        req.Name,
			Coordinates: coords,
			Active:      req.Active,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, areaToJSON(resp))
	}
}

// DeleteOperatingArea godoc
// @Summary Delete operating area
// @Tags settings
// @Security BearerAuth
// @Param id path string true "Area ID"
// @Success 200 {object} object
// @Router /settings/operating-areas/{id} [delete]
func DeleteOperatingArea(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx := getAuthContext(c)
		resp, err := clients.Settings.DeleteOperatingArea(ctx, &settingspb.DeleteOperatingAreaRequest{Id: id})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, areaToJSON(resp))
	}
}

type coord struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

func areaToJSON(a *settingspb.OperatingArea) gin.H {
	coords := make([]gin.H, len(a.Coordinates))
	for i, c := range a.Coordinates {
		coords[i] = gin.H{"lng": c.Lng, "lat": c.Lat}
	}
	return gin.H{
		"id":          a.Id,
		"name":        a.Name,
		"coordinates": coords,
		"active":      a.Active,
	}
}

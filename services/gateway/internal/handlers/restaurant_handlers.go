package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mendmzury/food-delivery/pkg/config"
	commonpb "github.com/mendmzury/food-delivery/proto/common"
	restaurantpb "github.com/mendmzury/food-delivery/proto/restaurant"
	"github.com/mendmzury/food-delivery/services/gateway/internal/grpc"
)

// ListRestaurants godoc
// @Summary List restaurants
// @Tags restaurants
// @Security BearerAuth
// @Success 200 {object} object
// @Router /restaurants [get]
func ListRestaurants(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, pageSize := getPaginationParams(c)
		search := c.Query("search")
		status := c.Query("status")

		ctx := getAuthContext(c)
		resp, err := clients.Restaurant.ListRestaurants(ctx, &restaurantpb.ListRestaurantsRequest{
			Pagination: &commonpb.PaginationRequest{Page: page, PageSize: pageSize},
			Search:     search,
			Status:     status,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		paginatedResponse(c, resp.Restaurants, resp.Pagination.Total, page, pageSize)
	}
}

// GetRestaurant godoc
// @Summary Get restaurant
// @Tags restaurants
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Success 200 {object} object
// @Router /restaurants/{id} [get]
func GetRestaurant(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx := getAuthContext(c)
		r, err := clients.Restaurant.GetRestaurant(ctx, &restaurantpb.GetRestaurantRequest{Id: id})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, r)
	}
}

// CreateRestaurant godoc
// @Summary Create restaurant
// @Tags restaurants
// @Security BearerAuth
// @Success 201 {object} object
// @Router /restaurants [post]
func CreateRestaurant(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name         string  `json:"name" binding:"required"`
			Description  string  `json:"description"`
			Logo         string  `json:"logo"`
			CoverImage   string  `json:"cover_image"`
			Address      string  `json:"address"`
			Lat          float64 `json:"lat"`
			Lng          float64 `json:"lng"`
			OpeningHours string  `json:"opening_hours"`
			OwnerID      string  `json:"owner_id"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		ctx := getAuthContext(c)
		r, err := clients.Restaurant.CreateRestaurant(ctx, &restaurantpb.CreateRestaurantRequest{
			Name:         req.Name,
			Description:  req.Description,
			Logo:         req.Logo,
			CoverImage:   req.CoverImage,
			Address:      req.Address,
			Lat:          req.Lat,
			Lng:          req.Lng,
			OpeningHours: req.OpeningHours,
			OwnerId:      req.OwnerID,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		createdResponse(c, r)
	}
}

// UpdateRestaurant godoc
// @Summary Update restaurant
// @Tags restaurants
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Success 200 {object} object
// @Router /restaurants/{id} [put]
func UpdateRestaurant(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Name         *string  `json:"name"`
			Description  *string  `json:"description"`
			Logo         *string  `json:"logo"`
			CoverImage   *string  `json:"cover_image"`
			Address      *string  `json:"address"`
			Lat          *float64 `json:"lat"`
			Lng          *float64 `json:"lng"`
			OpeningHours *string  `json:"opening_hours"`
			Status       *string  `json:"status"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		updateReq := &restaurantpb.UpdateRestaurantRequest{Id: id}
		if req.Name != nil {
			updateReq.Name = req.Name
		}
		if req.Description != nil {
			updateReq.Description = req.Description
		}
		if req.Logo != nil {
			updateReq.Logo = req.Logo
		}
		if req.CoverImage != nil {
			updateReq.CoverImage = req.CoverImage
		}
		if req.Address != nil {
			updateReq.Address = req.Address
		}
		if req.Lat != nil {
			updateReq.Lat = req.Lat
		}
		if req.Lng != nil {
			updateReq.Lng = req.Lng
		}
		if req.OpeningHours != nil {
			updateReq.OpeningHours = req.OpeningHours
		}
		if req.Status != nil {
			updateReq.Status = req.Status
		}
		ctx := getAuthContext(c)
		r, err := clients.Restaurant.UpdateRestaurant(ctx, updateReq)
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, r)
	}
}

// DeleteRestaurant godoc
// @Summary Delete restaurant
// @Tags restaurants
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Success 200 {object} object
// @Router /restaurants/{id} [delete]
func DeleteRestaurant(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		ctx := getAuthContext(c)
		_, err := clients.Restaurant.DeleteRestaurant(ctx, &restaurantpb.DeleteRestaurantRequest{Id: id})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, gin.H{"message": "deleted"})
	}
}

// ListCategories godoc
// @Summary List categories
// @Tags restaurants
// @Security BearerAuth
// @Param id path string true "Restaurant ID"
// @Success 200 {object} object
// @Router /restaurants/{id}/categories [get]
func ListCategories(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		restaurantID := c.Query("restaurant_id")
		if restaurantID == "" {
			restaurantID = c.Param("id")
		}
		if restaurantID == "" {
			errorResponse(c, http.StatusBadRequest, "restaurant_id required")
			return
		}
		ctx := getAuthContext(c)
		resp, err := clients.Restaurant.ListCategories(ctx, &restaurantpb.ListCategoriesRequest{RestaurantId: restaurantID})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, resp.Categories)
	}
}

// CreateCategory godoc
// @Summary Create category
// @Tags restaurants
// @Security BearerAuth
// @Success 201 {object} object
// @Router /restaurants/categories [post]
func CreateCategory(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name         string `json:"name" binding:"required"`
			RestaurantID string `json:"restaurant_id" binding:"required"`
			SortOrder    int32  `json:"sort_order"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		ctx := getAuthContext(c)
		cat, err := clients.Restaurant.CreateCategory(ctx, &restaurantpb.CreateCategoryRequest{
			Name:         req.Name,
			RestaurantId: req.RestaurantID,
			SortOrder:    req.SortOrder,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		createdResponse(c, cat)
	}
}

// ListMenuItems godoc
// @Summary List menu items
// @Tags restaurants
// @Security BearerAuth
// @Success 200 {object} object
// @Router /restaurants/menu-items [get]
func ListMenuItems(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryID := c.Query("category_id")
		restaurantID := c.Query("restaurant_id")
		availableOnly, _ := strconv.ParseBool(c.Query("available_only"))

		if categoryID == "" && restaurantID == "" {
			errorResponse(c, http.StatusBadRequest, "category_id or restaurant_id required")
			return
		}
		ctx := getAuthContext(c)
		resp, err := clients.Restaurant.ListMenuItems(ctx, &restaurantpb.ListMenuItemsRequest{
			CategoryId:    categoryID,
			RestaurantId:  restaurantID,
			AvailableOnly: availableOnly,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		successResponse(c, resp.MenuItems)
	}
}

// CreateMenuItem godoc
// @Summary Create menu item
// @Tags restaurants
// @Security BearerAuth
// @Success 201 {object} object
// @Router /restaurants/menu-items [post]
func CreateMenuItem(cfg *config.Config, clients *grpc.Clients) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string  `json:"name" binding:"required"`
			Description string  `json:"description"`
			Price       float64 `json:"price" binding:"required"`
			Image       string  `json:"image"`
			CategoryID  string  `json:"category_id" binding:"required"`
			Available   bool    `json:"available"`
			Options     string  `json:"options"`
			SortOrder   int32   `json:"sort_order"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "invalid request")
			return
		}
		ctx := getAuthContext(c)
		item, err := clients.Restaurant.CreateMenuItem(ctx, &restaurantpb.CreateMenuItemRequest{
			Name:        req.Name,
			Description: req.Description,
			Price:       req.Price,
			Image:       req.Image,
			CategoryId:  req.CategoryID,
			Available:   req.Available,
			Options:     req.Options,
			SortOrder:   req.SortOrder,
		})
		if err != nil {
			handleGRPCError(c, err)
			return
		}
		createdResponse(c, item)
	}
}

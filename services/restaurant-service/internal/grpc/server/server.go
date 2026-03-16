package server

import (
	"context"

	"github.com/mendmzury/food-delivery/services/restaurant-service/internal/application"
	restaurantpb "github.com/mendmzury/food-delivery/proto/restaurant"
	commonpb "github.com/mendmzury/food-delivery/proto/common"
)

type RestaurantServer struct {
	restaurantpb.UnimplementedRestaurantServiceServer
	service *application.RestaurantService
}

func NewRestaurantServer(service *application.RestaurantService) *RestaurantServer {
	return &RestaurantServer{service: service}
}

func (s *RestaurantServer) CreateRestaurant(ctx context.Context, req *restaurantpb.CreateRestaurantRequest) (*restaurantpb.Restaurant, error) {
	return s.service.CreateRestaurant(ctx, req)
}

func (s *RestaurantServer) GetRestaurant(ctx context.Context, req *restaurantpb.GetRestaurantRequest) (*restaurantpb.Restaurant, error) {
	return s.service.GetRestaurant(ctx, req)
}

func (s *RestaurantServer) ListRestaurants(ctx context.Context, req *restaurantpb.ListRestaurantsRequest) (*restaurantpb.ListRestaurantsResponse, error) {
	return s.service.ListRestaurants(ctx, req)
}

func (s *RestaurantServer) UpdateRestaurant(ctx context.Context, req *restaurantpb.UpdateRestaurantRequest) (*restaurantpb.Restaurant, error) {
	return s.service.UpdateRestaurant(ctx, req)
}

func (s *RestaurantServer) DeleteRestaurant(ctx context.Context, req *restaurantpb.DeleteRestaurantRequest) (*commonpb.SuccessResponse, error) {
	return s.service.DeleteRestaurant(ctx, req)
}

func (s *RestaurantServer) CreateCategory(ctx context.Context, req *restaurantpb.CreateCategoryRequest) (*restaurantpb.Category, error) {
	return s.service.CreateCategory(ctx, req)
}

func (s *RestaurantServer) GetCategory(ctx context.Context, req *restaurantpb.GetCategoryRequest) (*restaurantpb.Category, error) {
	return s.service.GetCategory(ctx, req)
}

func (s *RestaurantServer) ListCategories(ctx context.Context, req *restaurantpb.ListCategoriesRequest) (*restaurantpb.ListCategoriesResponse, error) {
	return s.service.ListCategories(ctx, req)
}

func (s *RestaurantServer) UpdateCategory(ctx context.Context, req *restaurantpb.UpdateCategoryRequest) (*restaurantpb.Category, error) {
	return s.service.UpdateCategory(ctx, req)
}

func (s *RestaurantServer) DeleteCategory(ctx context.Context, req *restaurantpb.DeleteCategoryRequest) (*commonpb.SuccessResponse, error) {
	return s.service.DeleteCategory(ctx, req)
}

func (s *RestaurantServer) CreateMenuItem(ctx context.Context, req *restaurantpb.CreateMenuItemRequest) (*restaurantpb.MenuItem, error) {
	return s.service.CreateMenuItem(ctx, req)
}

func (s *RestaurantServer) GetMenuItem(ctx context.Context, req *restaurantpb.GetMenuItemRequest) (*restaurantpb.MenuItem, error) {
	return s.service.GetMenuItem(ctx, req)
}

func (s *RestaurantServer) ListMenuItems(ctx context.Context, req *restaurantpb.ListMenuItemsRequest) (*restaurantpb.ListMenuItemsResponse, error) {
	return s.service.ListMenuItems(ctx, req)
}

func (s *RestaurantServer) UpdateMenuItem(ctx context.Context, req *restaurantpb.UpdateMenuItemRequest) (*restaurantpb.MenuItem, error) {
	return s.service.UpdateMenuItem(ctx, req)
}

func (s *RestaurantServer) DeleteMenuItem(ctx context.Context, req *restaurantpb.DeleteMenuItemRequest) (*commonpb.SuccessResponse, error) {
	return s.service.DeleteMenuItem(ctx, req)
}

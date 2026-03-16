package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/mendmzury/food-delivery/services/restaurant-service/internal/domain"
	"github.com/mendmzury/food-delivery/services/restaurant-service/internal/repository"
	restaurantpb "github.com/mendmzury/food-delivery/proto/restaurant"
	commonpb "github.com/mendmzury/food-delivery/proto/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RestaurantService struct {
	restaurantRepo *repository.RestaurantRepository
	categoryRepo   *repository.CategoryRepository
	menuItemRepo   *repository.MenuItemRepository
}

func NewRestaurantService(
	restaurantRepo *repository.RestaurantRepository,
	categoryRepo *repository.CategoryRepository,
	menuItemRepo *repository.MenuItemRepository,
) *RestaurantService {
	return &RestaurantService{
		restaurantRepo: restaurantRepo,
		categoryRepo:   categoryRepo,
		menuItemRepo:   menuItemRepo,
	}
}

func restaurantToProto(r *domain.Restaurant, ownerName string) *restaurantpb.Restaurant {
	if r == nil {
		return nil
	}
	return &restaurantpb.Restaurant{
		Id:           r.ID.String(),
		Name:         r.Name,
		Description:  r.Description,
		Logo:         r.Logo,
		CoverImage:   r.CoverImage,
		Address:      r.Address,
		Lat:          r.Lat,
		Lng:          r.Lng,
		OpeningHours: r.OpeningHours,
		Status:       r.Status,
		OwnerId:      r.OwnerID.String(),
		OwnerName:    ownerName,
		CreatedAt:    timestamppb.New(r.CreatedAt),
		UpdatedAt:    timestamppb.New(r.UpdatedAt),
	}
}

func categoryToProto(c *domain.Category) *restaurantpb.Category {
	if c == nil {
		return nil
	}
	return &restaurantpb.Category{
		Id:           c.ID.String(),
		Name:         c.Name,
		RestaurantId: c.RestaurantID.String(),
		SortOrder:    c.SortOrder,
		CreatedAt:    timestamppb.New(c.CreatedAt),
		UpdatedAt:    timestamppb.New(c.UpdatedAt),
	}
}

func menuItemToProto(m *domain.MenuItem) *restaurantpb.MenuItem {
	if m == nil {
		return nil
	}
	return &restaurantpb.MenuItem{
		Id:          m.ID.String(),
		Name:        m.Name,
		Description: m.Description,
		Price:       m.Price,
		Image:       m.Image,
		CategoryId:  m.CategoryID.String(),
		Available:   m.Available,
		Options:     m.Options,
		SortOrder:   m.SortOrder,
		CreatedAt:   timestamppb.New(m.CreatedAt),
		UpdatedAt:   timestamppb.New(m.UpdatedAt),
	}
}

func (s *RestaurantService) CreateRestaurant(ctx context.Context, req *restaurantpb.CreateRestaurantRequest) (*restaurantpb.Restaurant, error) {
	ownerID := uuid.Nil
	if req.OwnerId != "" {
		var err error
		ownerID, err = uuid.Parse(req.OwnerId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
		}
	}

	r := &domain.Restaurant{
		Name:         req.Name,
		Description:  req.Description,
		Logo:         req.Logo,
		CoverImage:   req.CoverImage,
		Address:      req.Address,
		Lat:          req.Lat,
		Lng:          req.Lng,
		OpeningHours: req.OpeningHours,
		Status:       "active",
		OwnerID:      ownerID,
	}
	if err := s.restaurantRepo.Create(ctx, r); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return restaurantToProto(r, ""), nil
}

func (s *RestaurantService) GetRestaurant(ctx context.Context, req *restaurantpb.GetRestaurantRequest) (*restaurantpb.Restaurant, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	r, err := s.restaurantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if r == nil {
		return nil, status.Error(codes.NotFound, "restaurant not found")
	}
	return restaurantToProto(r, ""), nil
}

func (s *RestaurantService) ListRestaurants(ctx context.Context, req *restaurantpb.ListRestaurantsRequest) (*restaurantpb.ListRestaurantsResponse, error) {
	page, pageSize := int64(1), int64(10)
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
		if req.Pagination.PageSize > 0 {
			pageSize = req.Pagination.PageSize
		}
	}
	if pageSize > 100 {
		pageSize = 100
	}

	restaurants, total, err := s.restaurantRepo.List(ctx, req.Search, req.Status, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoRestaurants := make([]*restaurantpb.Restaurant, len(restaurants))
	for i, r := range restaurants {
		protoRestaurants[i] = restaurantToProto(r, "")
	}

	totalPages := int64(0)
	if total > 0 && pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return &restaurantpb.ListRestaurantsResponse{
		Restaurants: protoRestaurants,
		Pagination: &commonpb.PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}, nil
}

func (s *RestaurantService) UpdateRestaurant(ctx context.Context, req *restaurantpb.UpdateRestaurantRequest) (*restaurantpb.Restaurant, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	r, err := s.restaurantRepo.GetByID(ctx, id)
	if err != nil || r == nil {
		return nil, status.Error(codes.NotFound, "restaurant not found")
	}

	if req.Name != nil {
		r.Name = *req.Name
	}
	if req.Description != nil {
		r.Description = *req.Description
	}
	if req.Logo != nil {
		r.Logo = *req.Logo
	}
	if req.CoverImage != nil {
		r.CoverImage = *req.CoverImage
	}
	if req.Address != nil {
		r.Address = *req.Address
	}
	if req.Lat != nil {
		r.Lat = *req.Lat
	}
	if req.Lng != nil {
		r.Lng = *req.Lng
	}
	if req.OpeningHours != nil {
		r.OpeningHours = *req.OpeningHours
	}
	if req.Status != nil {
		r.Status = *req.Status
	}

	if err := s.restaurantRepo.Update(ctx, r); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return restaurantToProto(r, ""), nil
}

func (s *RestaurantService) DeleteRestaurant(ctx context.Context, req *restaurantpb.DeleteRestaurantRequest) (*commonpb.SuccessResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	if err := s.restaurantRepo.Delete(ctx, id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &commonpb.SuccessResponse{Success: true}, nil
}

func (s *RestaurantService) CreateCategory(ctx context.Context, req *restaurantpb.CreateCategoryRequest) (*restaurantpb.Category, error) {
	restaurantID, err := uuid.Parse(req.RestaurantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid restaurant_id")
	}
	c := &domain.Category{
		Name:         req.Name,
		RestaurantID: restaurantID,
		SortOrder:    req.SortOrder,
	}
	if err := s.categoryRepo.Create(ctx, c); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return categoryToProto(c), nil
}

func (s *RestaurantService) GetCategory(ctx context.Context, req *restaurantpb.GetCategoryRequest) (*restaurantpb.Category, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	c, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if c == nil {
		return nil, status.Error(codes.NotFound, "category not found")
	}
	return categoryToProto(c), nil
}

func (s *RestaurantService) ListCategories(ctx context.Context, req *restaurantpb.ListCategoriesRequest) (*restaurantpb.ListCategoriesResponse, error) {
	restaurantID, err := uuid.Parse(req.RestaurantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid restaurant_id")
	}
	categories, err := s.categoryRepo.ListByRestaurant(ctx, restaurantID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	protoCategories := make([]*restaurantpb.Category, len(categories))
	for i, c := range categories {
		protoCategories[i] = categoryToProto(c)
	}
	return &restaurantpb.ListCategoriesResponse{Categories: protoCategories}, nil
}

func (s *RestaurantService) UpdateCategory(ctx context.Context, req *restaurantpb.UpdateCategoryRequest) (*restaurantpb.Category, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	c, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil || c == nil {
		return nil, status.Error(codes.NotFound, "category not found")
	}
	if req.Name != nil {
		c.Name = *req.Name
	}
	if req.SortOrder != nil {
		c.SortOrder = *req.SortOrder
	}
	if err := s.categoryRepo.Update(ctx, c); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return categoryToProto(c), nil
}

func (s *RestaurantService) DeleteCategory(ctx context.Context, req *restaurantpb.DeleteCategoryRequest) (*commonpb.SuccessResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &commonpb.SuccessResponse{Success: true}, nil
}

func (s *RestaurantService) CreateMenuItem(ctx context.Context, req *restaurantpb.CreateMenuItemRequest) (*restaurantpb.MenuItem, error) {
	categoryID, err := uuid.Parse(req.CategoryId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid category_id")
	}
	m := &domain.MenuItem{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Image:       req.Image,
		CategoryID:  categoryID,
		Available:   req.Available,
		Options:     req.Options,
		SortOrder:   req.SortOrder,
	}
	if err := s.menuItemRepo.Create(ctx, m); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return menuItemToProto(m), nil
}

func (s *RestaurantService) GetMenuItem(ctx context.Context, req *restaurantpb.GetMenuItemRequest) (*restaurantpb.MenuItem, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	m, err := s.menuItemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if m == nil {
		return nil, status.Error(codes.NotFound, "menu item not found")
	}
	return menuItemToProto(m), nil
}

func (s *RestaurantService) ListMenuItems(ctx context.Context, req *restaurantpb.ListMenuItemsRequest) (*restaurantpb.ListMenuItemsResponse, error) {
	var items []*domain.MenuItem
	var err error

	if req.CategoryId != "" {
		categoryID, parseErr := uuid.Parse(req.CategoryId)
		if parseErr != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid category_id")
		}
		items, err = s.menuItemRepo.ListByCategory(ctx, categoryID, req.AvailableOnly)
	} else if req.RestaurantId != "" {
		restaurantID, parseErr := uuid.Parse(req.RestaurantId)
		if parseErr != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid restaurant_id")
		}
		items, err = s.menuItemRepo.ListByRestaurant(ctx, restaurantID, req.AvailableOnly)
	} else {
		return nil, status.Error(codes.InvalidArgument, "category_id or restaurant_id required")
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoItems := make([]*restaurantpb.MenuItem, len(items))
	for i, m := range items {
		protoItems[i] = menuItemToProto(m)
	}
	return &restaurantpb.ListMenuItemsResponse{MenuItems: protoItems}, nil
}

func (s *RestaurantService) UpdateMenuItem(ctx context.Context, req *restaurantpb.UpdateMenuItemRequest) (*restaurantpb.MenuItem, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	m, err := s.menuItemRepo.GetByID(ctx, id)
	if err != nil || m == nil {
		return nil, status.Error(codes.NotFound, "menu item not found")
	}
	if req.Name != nil {
		m.Name = *req.Name
	}
	if req.Description != nil {
		m.Description = *req.Description
	}
	if req.Price != nil {
		m.Price = *req.Price
	}
	if req.Image != nil {
		m.Image = *req.Image
	}
	if req.Available != nil {
		m.Available = *req.Available
	}
	if req.Options != nil {
		m.Options = *req.Options
	}
	if req.SortOrder != nil {
		m.SortOrder = *req.SortOrder
	}
	if err := s.menuItemRepo.Update(ctx, m); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return menuItemToProto(m), nil
}

func (s *RestaurantService) DeleteMenuItem(ctx context.Context, req *restaurantpb.DeleteMenuItemRequest) (*commonpb.SuccessResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}
	if err := s.menuItemRepo.Delete(ctx, id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &commonpb.SuccessResponse{Success: true}, nil
}

package app

import (
	"errors"
	"fmt"
	"time"

	"glovo-backend/services/catalog-service/internal/domain"

	"github.com/google/uuid"
)

type catalogService struct {
	storeRepo    domain.StoreRepository
	productRepo  domain.ProductRepository
	categoryRepo domain.CategoryRepository
}

func NewCatalogService(
	storeRepo domain.StoreRepository,
	productRepo domain.ProductRepository,
	categoryRepo domain.CategoryRepository,
) domain.CatalogService {
	return &catalogService{
		storeRepo:    storeRepo,
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

// Store management
func (s *catalogService) CreateStore(merchantID string, req domain.CreateStoreRequest) (*domain.Store, error) {
	// Check if merchant already has a store
	existingStore, _ := s.storeRepo.GetByMerchantID(merchantID)
	if existingStore != nil {
		return nil, errors.New("merchant already has a store")
	}

	store := &domain.Store{
		ID:           uuid.New().String(),
		MerchantID:   merchantID,
		Name:         req.Name,
		Description:  req.Description,
		Address:      req.Address,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		Phone:        req.Phone,
		Email:        req.Email,
		Status:       domain.StatusOpen,
		OpeningHours: req.OpeningHours,
		DeliveryInfo: req.DeliveryInfo,
		Rating:       0.0,
		ReviewCount:  0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Add categories if provided
	if len(req.CategoryIDs) > 0 {
		for _, categoryID := range req.CategoryIDs {
			category, err := s.categoryRepo.GetByID(categoryID)
			if err == nil && category != nil {
				store.Categories = append(store.Categories, *category)
			}
		}
	}

	if err := s.storeRepo.Create(store); err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	return store, nil
}

func (s *catalogService) GetStore(storeID string) (*domain.Store, error) {
	return s.storeRepo.GetByID(storeID)
}

func (s *catalogService) GetMerchantStore(merchantID string) (*domain.Store, error) {
	return s.storeRepo.GetByMerchantID(merchantID)
}

func (s *catalogService) UpdateStore(storeID string, merchantID string, updates map[string]interface{}) (*domain.Store, error) {
	store, err := s.storeRepo.GetByID(storeID)
	if err != nil {
		return nil, err
	}

	// Check if merchant owns this store
	if store.MerchantID != merchantID {
		return nil, errors.New("unauthorized to update this store")
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		store.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		store.Description = description
	}
	if status, ok := updates["status"].(string); ok {
		store.Status = domain.StoreStatus(status)
	}
	if phone, ok := updates["phone"].(string); ok {
		store.Phone = phone
	}
	if email, ok := updates["email"].(string); ok {
		store.Email = email
	}

	store.UpdatedAt = time.Now()

	if err := s.storeRepo.Update(store); err != nil {
		return nil, fmt.Errorf("failed to update store: %w", err)
	}

	return store, nil
}

func (s *catalogService) SearchStores(req domain.StoreSearchRequest) ([]domain.Store, error) {
	return s.storeRepo.Search(req)
}

// Product management
func (s *catalogService) CreateProduct(storeID string, merchantID string, req domain.CreateProductRequest) (*domain.Product, error) {
	// Verify store ownership
	store, err := s.storeRepo.GetByID(storeID)
	if err != nil {
		return nil, err
	}

	if store.MerchantID != merchantID {
		return nil, errors.New("unauthorized to add products to this store")
	}

	// Verify category exists
	_, err = s.categoryRepo.GetByID(req.CategoryID)
	if err != nil {
		return nil, errors.New("category not found")
	}

	product := &domain.Product{
		ID:          uuid.New().String(),
		StoreID:     storeID,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Image:       req.Image,
		Status:      domain.ProductStatusAvailable,
		Nutrition:   req.Nutrition,
		Tags:        req.Tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Add options if provided
	for _, optionReq := range req.Options {
		option := domain.ProductOption{
			ID:        uuid.New().String(),
			ProductID: product.ID,
			Name:      optionReq.Name,
			Type:      optionReq.Type,
			Required:  optionReq.Required,
		}

		for _, choiceReq := range optionReq.Options {
			choice := domain.ProductOptionChoice{
				ID:         uuid.New().String(),
				OptionID:   option.ID,
				Name:       choiceReq.Name,
				PriceExtra: choiceReq.PriceExtra,
			}
			option.Options = append(option.Options, choice)
		}

		product.Options = append(product.Options, option)
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

func (s *catalogService) GetProduct(productID string) (*domain.Product, error) {
	return s.productRepo.GetByID(productID)
}

func (s *catalogService) GetStoreProducts(storeID string, limit, offset int) ([]domain.Product, error) {
	return s.productRepo.GetByStoreID(storeID, limit, offset)
}

func (s *catalogService) UpdateProduct(productID string, merchantID string, updates map[string]interface{}) (*domain.Product, error) {
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		return nil, err
	}

	// Verify ownership through store
	store, err := s.storeRepo.GetByID(product.StoreID)
	if err != nil {
		return nil, err
	}

	if store.MerchantID != merchantID {
		return nil, errors.New("unauthorized to update this product")
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		product.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		product.Description = description
	}
	if price, ok := updates["price"].(float64); ok {
		product.Price = price
	}
	if status, ok := updates["status"].(string); ok {
		product.Status = domain.ProductStatus(status)
	}
	if image, ok := updates["image"].(string); ok {
		product.Image = image
	}

	product.UpdatedAt = time.Now()

	if err := s.productRepo.Update(product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

func (s *catalogService) DeleteProduct(productID string, merchantID string) error {
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		return err
	}

	// Verify ownership through store
	store, err := s.storeRepo.GetByID(product.StoreID)
	if err != nil {
		return err
	}

	if store.MerchantID != merchantID {
		return errors.New("unauthorized to delete this product")
	}

	return s.productRepo.Delete(productID)
}

func (s *catalogService) SearchProducts(query string, storeID string, limit, offset int) ([]domain.Product, error) {
	return s.productRepo.Search(query, storeID, limit, offset)
}

// Category management
func (s *catalogService) CreateCategory(req domain.CreateCategoryRequest) (*domain.Category, error) {
	// Verify parent category exists if provided
	if req.ParentID != nil {
		_, err := s.categoryRepo.GetByID(*req.ParentID)
		if err != nil {
			return nil, errors.New("parent category not found")
		}
	}

	category := &domain.Category{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Image:       req.Image,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

func (s *catalogService) GetCategory(categoryID string) (*domain.Category, error) {
	return s.categoryRepo.GetByID(categoryID)
}

func (s *catalogService) GetCategories() ([]domain.Category, error) {
	return s.categoryRepo.GetAll()
}

func (s *catalogService) UpdateCategory(categoryID string, updates map[string]interface{}) (*domain.Category, error) {
	category, err := s.categoryRepo.GetByID(categoryID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		category.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		category.Description = description
	}
	if image, ok := updates["image"].(string); ok {
		category.Image = image
	}
	if sortOrder, ok := updates["sort_order"].(int); ok {
		category.SortOrder = sortOrder
	}

	category.UpdatedAt = time.Now()

	if err := s.categoryRepo.Update(category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

func (s *catalogService) DeleteCategory(categoryID string) error {
	// Check if category has children
	children, err := s.categoryRepo.GetByParentID(&categoryID)
	if err == nil && len(children) > 0 {
		return errors.New("cannot delete category with subcategories")
	}

	// Check if category has products
	products, err := s.productRepo.GetByCategoryID(categoryID, 1, 0)
	if err == nil && len(products) > 0 {
		return errors.New("cannot delete category with products")
	}

	return s.categoryRepo.Delete(categoryID)
}

// Order validation (for Order Service)
func (s *catalogService) ValidateOrderItems(storeID string, items []domain.OrderItem) (*domain.OrderValidation, error) {
	var validatedItems []domain.ValidatedOrderItem
	var totalAmount float64
	var errors []string

	// Verify store exists and is open
	store, err := s.storeRepo.GetByID(storeID)
	if err != nil {
		return &domain.OrderValidation{
			Valid:  false,
			Errors: []string{"Store not found"},
		}, nil
	}

	if store.Status != domain.StatusOpen {
		return &domain.OrderValidation{
			Valid:  false,
			Errors: []string{"Store is currently closed"},
		}, nil
	}

	for _, item := range items {
		product, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Product %s not found", item.ProductID))
			continue
		}

		if product.StoreID != storeID {
			errors = append(errors, fmt.Sprintf("Product %s does not belong to this store", item.ProductID))
			continue
		}

		available := product.Status == domain.ProductStatusAvailable
		subtotal := product.Price * float64(item.Quantity)

		validatedItem := domain.ValidatedOrderItem{
			ProductID: item.ProductID,
			Name:      product.Name,
			Price:     product.Price,
			Quantity:  item.Quantity,
			Available: available,
			Subtotal:  subtotal,
		}

		validatedItems = append(validatedItems, validatedItem)

		if available {
			totalAmount += subtotal
		} else {
			errors = append(errors, fmt.Sprintf("Product %s is not available", product.Name))
		}
	}

	// Check minimum order amount
	if totalAmount > 0 && totalAmount < store.DeliveryInfo.MinOrderAmount {
		errors = append(errors, fmt.Sprintf("Minimum order amount is $%.2f", store.DeliveryInfo.MinOrderAmount))
	}

	return &domain.OrderValidation{
		Valid:       len(errors) == 0,
		Items:       validatedItems,
		TotalAmount: totalAmount,
		Errors:      errors,
	}, nil
}

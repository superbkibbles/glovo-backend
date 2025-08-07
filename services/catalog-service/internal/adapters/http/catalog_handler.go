package http

import (
	"net/http"
	"strconv"

	"glovo-backend/services/catalog-service/internal/domain"
	"glovo-backend/shared/auth"
	"glovo-backend/shared/middleware"

	"github.com/gin-gonic/gin"
)

type CatalogHandler struct {
	catalogService domain.CatalogService
}

func NewCatalogHandler(catalogService domain.CatalogService) *CatalogHandler {
	return &CatalogHandler{catalogService: catalogService}
}

func (h *CatalogHandler) SetupRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		// Public routes
		v1.GET("/stores", h.SearchStores)
		v1.GET("/stores/:id", h.GetStore)
		v1.GET("/stores/:id/products", h.GetStoreProducts)
		v1.GET("/products/:id", h.GetProduct)
		v1.GET("/categories", h.GetCategories)
		v1.GET("/categories/:id", h.GetCategory)
		v1.POST("/stores/:id/validate-order", h.ValidateOrder)

		// Merchant routes
		merchant := v1.Group("/merchant")
		merchant.Use(middleware.AuthMiddleware())
		merchant.Use(middleware.RequireRoles([]auth.UserRole{auth.RoleMerchant, auth.RoleAdmin}))
		{
			merchant.POST("/store", h.CreateStore)
			merchant.GET("/store", h.GetMerchantStore)
			merchant.PUT("/store", h.UpdateStore)
			merchant.POST("/store/products", h.CreateProduct)
			merchant.PUT("/products/:id", h.UpdateProduct)
			merchant.DELETE("/products/:id", h.DeleteProduct)
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.RequireRole(auth.RoleAdmin))
		{
			admin.POST("/categories", h.CreateCategory)
			admin.PUT("/categories/:id", h.UpdateCategory)
			admin.DELETE("/categories/:id", h.DeleteCategory)
			admin.GET("/stores", h.GetAllStores)
		}
	}
}

// SearchStores godoc
// @Summary Search stores
// @Description Search for stores based on location, category, and other filters
// @Tags Stores
// @Produce json
// @Param query query string false "Search query"
// @Param category_id query string false "Category ID"
// @Param latitude query number false "User latitude"
// @Param longitude query number false "User longitude"
// @Param radius query number false "Search radius in km"
// @Param min_rating query number false "Minimum rating"
// @Param sort_by query string false "Sort by: rating, distance, delivery_time"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.Store
// @Failure 500 {object} map[string]string
// @Router /api/v1/stores [get]
func (h *CatalogHandler) SearchStores(c *gin.Context) {
	req := domain.StoreSearchRequest{
		Query:      c.Query("query"),
		CategoryID: c.Query("category_id"),
		SortBy:     c.Query("sort_by"),
	}

	if lat := c.Query("latitude"); lat != "" {
		if latFloat, err := strconv.ParseFloat(lat, 64); err == nil {
			req.Latitude = latFloat
		}
	}

	if lng := c.Query("longitude"); lng != "" {
		if lngFloat, err := strconv.ParseFloat(lng, 64); err == nil {
			req.Longitude = lngFloat
		}
	}

	if radius := c.Query("radius"); radius != "" {
		if radiusFloat, err := strconv.ParseFloat(radius, 64); err == nil {
			req.Radius = radiusFloat
		}
	}

	if rating := c.Query("min_rating"); rating != "" {
		if ratingFloat, err := strconv.ParseFloat(rating, 64); err == nil {
			req.MinRating = ratingFloat
		}
	}

	req.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))
	req.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))

	stores, err := h.catalogService.SearchStores(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stores)
}

// GetStore godoc
// @Summary Get store by ID
// @Description Get detailed information about a store
// @Tags Stores
// @Produce json
// @Param id path string true "Store ID"
// @Success 200 {object} domain.Store
// @Failure 404 {object} map[string]string
// @Router /api/v1/stores/{id} [get]
func (h *CatalogHandler) GetStore(c *gin.Context) {
	storeID := c.Param("id")

	store, err := h.catalogService.GetStore(storeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	c.JSON(http.StatusOK, store)
}

// CreateStore godoc
// @Summary Create a new store
// @Description Create a new store for the authenticated merchant
// @Tags Merchant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateStoreRequest true "Store details"
// @Success 201 {object} domain.Store
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/merchant/store [post]
func (h *CatalogHandler) CreateStore(c *gin.Context) {
	merchantID := c.GetString("user_id")
	if merchantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Merchant ID not found"})
		return
	}

	var req domain.CreateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	store, err := h.catalogService.CreateStore(merchantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, store)
}

// GetMerchantStore godoc
// @Summary Get merchant's store
// @Description Get the store belonging to the authenticated merchant
// @Tags Merchant
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Store
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/merchant/store [get]
func (h *CatalogHandler) GetMerchantStore(c *gin.Context) {
	merchantID := c.GetString("user_id")

	store, err := h.catalogService.GetMerchantStore(merchantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	c.JSON(http.StatusOK, store)
}

// UpdateStore godoc
// @Summary Update store
// @Description Update store information
// @Tags Merchant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param updates body map[string]interface{} true "Store updates"
// @Success 200 {object} domain.Store
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/merchant/store [put]
func (h *CatalogHandler) UpdateStore(c *gin.Context) {
	merchantID := c.GetString("user_id")

	// First get the store to get the store ID
	store, err := h.catalogService.GetMerchantStore(merchantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedStore, err := h.catalogService.UpdateStore(store.ID, merchantID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedStore)
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Add a new product to the merchant's store
// @Tags Merchant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateProductRequest true "Product details"
// @Success 201 {object} domain.Product
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/merchant/store/products [post]
func (h *CatalogHandler) CreateProduct(c *gin.Context) {
	merchantID := c.GetString("user_id")

	// Get the merchant's store
	store, err := h.catalogService.GetMerchantStore(merchantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	var req domain.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.catalogService.CreateProduct(store.ID, merchantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GetStoreProducts godoc
// @Summary Get store products
// @Description Get all products for a specific store
// @Tags Stores
// @Produce json
// @Param id path string true "Store ID"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.Product
// @Failure 404 {object} map[string]string
// @Router /api/v1/stores/{id}/products [get]
func (h *CatalogHandler) GetStoreProducts(c *gin.Context) {
	storeID := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	products, err := h.catalogService.GetStoreProducts(storeID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetProduct godoc
// @Summary Get product by ID
// @Description Get detailed information about a product
// @Tags Products
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} domain.Product
// @Failure 404 {object} map[string]string
// @Router /api/v1/products/{id} [get]
func (h *CatalogHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")

	product, err := h.catalogService.GetProduct(productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// UpdateProduct godoc
// @Summary Update product
// @Description Update product information
// @Tags Merchant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param updates body map[string]interface{} true "Product updates"
// @Success 200 {object} domain.Product
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/merchant/products/{id} [put]
func (h *CatalogHandler) UpdateProduct(c *gin.Context) {
	productID := c.Param("id")
	merchantID := c.GetString("user_id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.catalogService.UpdateProduct(productID, merchantID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct godoc
// @Summary Delete product
// @Description Delete a product from the store
// @Tags Merchant
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/merchant/products/{id} [delete]
func (h *CatalogHandler) DeleteProduct(c *gin.Context) {
	productID := c.Param("id")
	merchantID := c.GetString("user_id")

	err := h.catalogService.DeleteProduct(productID, merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// GetCategories godoc
// @Summary Get all categories
// @Description Get all product categories
// @Tags Categories
// @Produce json
// @Success 200 {array} domain.Category
// @Failure 500 {object} map[string]string
// @Router /api/v1/categories [get]
func (h *CatalogHandler) GetCategories(c *gin.Context) {
	categories, err := h.catalogService.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCategory godoc
// @Summary Get category by ID
// @Description Get detailed information about a category
// @Tags Categories
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} domain.Category
// @Failure 404 {object} map[string]string
// @Router /api/v1/categories/{id} [get]
func (h *CatalogHandler) GetCategory(c *gin.Context) {
	categoryID := c.Param("id")

	category, err := h.catalogService.GetCategory(categoryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// CreateCategory godoc
// @Summary Create a new category
// @Description Create a new product category (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateCategoryRequest true "Category details"
// @Success 201 {object} domain.Category
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/admin/categories [post]
func (h *CatalogHandler) CreateCategory(c *gin.Context) {
	var req domain.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.catalogService.CreateCategory(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// UpdateCategory godoc
// @Summary Update category
// @Description Update category information (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param updates body map[string]interface{} true "Category updates"
// @Success 200 {object} domain.Category
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/admin/categories/{id} [put]
func (h *CatalogHandler) UpdateCategory(c *gin.Context) {
	categoryID := c.Param("id")

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.catalogService.UpdateCategory(categoryID, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory godoc
// @Summary Delete category
// @Description Delete a category (Admin only)
// @Tags Admin
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/admin/categories/{id} [delete]
func (h *CatalogHandler) DeleteCategory(c *gin.Context) {
	categoryID := c.Param("id")

	err := h.catalogService.DeleteCategory(categoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

// ValidateOrder godoc
// @Summary Validate order items
// @Description Validate order items for a specific store (used by Order Service)
// @Tags Stores
// @Accept json
// @Produce json
// @Param id path string true "Store ID"
// @Param items body []domain.OrderItem true "Order items to validate"
// @Success 200 {object} domain.OrderValidation
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/stores/{id}/validate-order [post]
func (h *CatalogHandler) ValidateOrder(c *gin.Context) {
	storeID := c.Param("id")

	var items []domain.OrderItem
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validation, err := h.catalogService.ValidateOrderItems(storeID, items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, validation)
}

// GetAllStores godoc
// @Summary Get all stores (Admin only)
// @Description Get all stores in the system with pagination
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} domain.Store
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/v1/admin/stores [get]
func (h *CatalogHandler) GetAllStores(c *gin.Context) {
	// For admin, use search with no filters to get all stores
	req := domain.StoreSearchRequest{}
	req.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))
	req.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))

	stores, err := h.catalogService.SearchStores(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stores)
}

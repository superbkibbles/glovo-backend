package domain

import (
	"time"
)

// Store represents a merchant's store/restaurant
type Store struct {
	ID           string       `json:"id" gorm:"primaryKey"`
	MerchantID   string       `json:"merchant_id" gorm:"index"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Address      string       `json:"address"`
	Latitude     float64      `json:"latitude"`
	Longitude    float64      `json:"longitude"`
	Phone        string       `json:"phone"`
	Email        string       `json:"email"`
	Status       StoreStatus  `json:"status"`
	Categories   []Category   `json:"categories" gorm:"many2many:store_categories;"`
	Products     []Product    `json:"products" gorm:"foreignKey:StoreID"`
	Rating       float64      `json:"rating"`
	ReviewCount  int          `json:"review_count"`
	OpeningHours OpeningHours `json:"opening_hours" gorm:"embedded"`
	DeliveryInfo DeliveryInfo `json:"delivery_info" gorm:"embedded"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type StoreStatus string

const (
	StatusOpen   StoreStatus = "open"
	StatusClosed StoreStatus = "closed"
	StatusPaused StoreStatus = "paused"
)

type OpeningHours struct {
	Monday    string `json:"monday"`
	Tuesday   string `json:"tuesday"`
	Wednesday string `json:"wednesday"`
	Thursday  string `json:"thursday"`
	Friday    string `json:"friday"`
	Saturday  string `json:"saturday"`
	Sunday    string `json:"sunday"`
}

type DeliveryInfo struct {
	MinOrderAmount float64 `json:"min_order_amount"`
	DeliveryFee    float64 `json:"delivery_fee"`
	DeliveryRadius float64 `json:"delivery_radius"` // in kilometers
	EstimatedTime  int     `json:"estimated_time"`  // in minutes
}

// Product represents an item that can be ordered
type Product struct {
	ID          string          `json:"id" gorm:"primaryKey"`
	StoreID     string          `json:"store_id" gorm:"index"`
	CategoryID  string          `json:"category_id" gorm:"index"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       float64         `json:"price"`
	Image       string          `json:"image,omitempty"`
	Status      ProductStatus   `json:"status"`
	Options     []ProductOption `json:"options" gorm:"foreignKey:ProductID"`
	Nutrition   NutritionInfo   `json:"nutrition" gorm:"embedded"`
	Tags        []string        `json:"tags" gorm:"serializer:json"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ProductStatus string

const (
	ProductStatusAvailable   ProductStatus = "available"
	ProductStatusUnavailable ProductStatus = "unavailable"
	ProductStatusSoldOut     ProductStatus = "sold_out"
)

type ProductOption struct {
	ID        string                `json:"id" gorm:"primaryKey"`
	ProductID string                `json:"product_id"`
	Name      string                `json:"name"`
	Type      ProductOptionType     `json:"type"`
	Required  bool                  `json:"required"`
	Options   []ProductOptionChoice `json:"options" gorm:"foreignKey:OptionID"`
}

type ProductOptionType string

const (
	OptionTypeSingle   ProductOptionType = "single"   // radio buttons
	OptionTypeMultiple ProductOptionType = "multiple" // checkboxes
)

type ProductOptionChoice struct {
	ID         string  `json:"id" gorm:"primaryKey"`
	OptionID   string  `json:"option_id"`
	Name       string  `json:"name"`
	PriceExtra float64 `json:"price_extra"`
}

type NutritionInfo struct {
	Calories int     `json:"calories,omitempty"`
	Protein  float64 `json:"protein,omitempty"`
	Carbs    float64 `json:"carbs,omitempty"`
	Fat      float64 `json:"fat,omitempty"`
	Fiber    float64 `json:"fiber,omitempty"`
	Sugar    float64 `json:"sugar,omitempty"`
	Sodium   float64 `json:"sodium,omitempty"`
}

// Category represents product categories
type Category struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Image       string     `json:"image,omitempty"`
	ParentID    *string    `json:"parent_id,omitempty"`
	Children    []Category `json:"children" gorm:"foreignKey:ParentID"`
	SortOrder   int        `json:"sort_order"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Request/Response DTOs
type CreateStoreRequest struct {
	Name         string       `json:"name" binding:"required"`
	Description  string       `json:"description"`
	Address      string       `json:"address" binding:"required"`
	Latitude     float64      `json:"latitude" binding:"required"`
	Longitude    float64      `json:"longitude" binding:"required"`
	Phone        string       `json:"phone" binding:"required"`
	Email        string       `json:"email" binding:"required,email"`
	CategoryIDs  []string     `json:"category_ids"`
	OpeningHours OpeningHours `json:"opening_hours"`
	DeliveryInfo DeliveryInfo `json:"delivery_info"`
}

type CreateProductRequest struct {
	CategoryID  string             `json:"category_id" binding:"required"`
	Name        string             `json:"name" binding:"required"`
	Description string             `json:"description"`
	Price       float64            `json:"price" binding:"required,min=0"`
	Image       string             `json:"image"`
	Options     []ProductOptionReq `json:"options"`
	Nutrition   NutritionInfo      `json:"nutrition"`
	Tags        []string           `json:"tags"`
}

type ProductOptionReq struct {
	Name     string                   `json:"name" binding:"required"`
	Type     ProductOptionType        `json:"type" binding:"required"`
	Required bool                     `json:"required"`
	Options  []ProductOptionChoiceReq `json:"options" binding:"required,min=1"`
}

type ProductOptionChoiceReq struct {
	Name       string  `json:"name" binding:"required"`
	PriceExtra float64 `json:"price_extra"`
}

type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	ParentID    *string `json:"parent_id"`
	SortOrder   int     `json:"sort_order"`
}

type StoreSearchRequest struct {
	Query      string  `json:"query,omitempty"`
	CategoryID string  `json:"category_id,omitempty"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
	Radius     float64 `json:"radius,omitempty"` // in kilometers
	MinRating  float64 `json:"min_rating,omitempty"`
	SortBy     string  `json:"sort_by,omitempty"` // rating, distance, delivery_time
	Limit      int     `json:"limit,omitempty"`
	Offset     int     `json:"offset,omitempty"`
}

// Repository interfaces (ports)
type StoreRepository interface {
	Create(store *Store) error
	GetByID(id string) (*Store, error)
	GetByMerchantID(merchantID string) (*Store, error)
	Update(store *Store) error
	Delete(id string) error
	Search(req StoreSearchRequest) ([]Store, error)
	List(limit, offset int) ([]Store, error)
}

type ProductRepository interface {
	Create(product *Product) error
	GetByID(id string) (*Product, error)
	GetByStoreID(storeID string, limit, offset int) ([]Product, error)
	GetByCategoryID(categoryID string, limit, offset int) ([]Product, error)
	Update(product *Product) error
	Delete(id string) error
	Search(query string, storeID string, limit, offset int) ([]Product, error)
}

type CategoryRepository interface {
	Create(category *Category) error
	GetByID(id string) (*Category, error)
	GetAll() ([]Category, error)
	GetByParentID(parentID *string) ([]Category, error)
	Update(category *Category) error
	Delete(id string) error
}

// Service interfaces (ports)
type CatalogService interface {
	// Store management
	CreateStore(merchantID string, req CreateStoreRequest) (*Store, error)
	GetStore(storeID string) (*Store, error)
	GetMerchantStore(merchantID string) (*Store, error)
	UpdateStore(storeID string, merchantID string, updates map[string]interface{}) (*Store, error)
	SearchStores(req StoreSearchRequest) ([]Store, error)

	// Product management
	CreateProduct(storeID string, merchantID string, req CreateProductRequest) (*Product, error)
	GetProduct(productID string) (*Product, error)
	GetStoreProducts(storeID string, limit, offset int) ([]Product, error)
	UpdateProduct(productID string, merchantID string, updates map[string]interface{}) (*Product, error)
	DeleteProduct(productID string, merchantID string) error
	SearchProducts(query string, storeID string, limit, offset int) ([]Product, error)

	// Category management
	CreateCategory(req CreateCategoryRequest) (*Category, error)
	GetCategory(categoryID string) (*Category, error)
	GetCategories() ([]Category, error)
	UpdateCategory(categoryID string, updates map[string]interface{}) (*Category, error)
	DeleteCategory(categoryID string) error

	// Order validation (for Order Service)
	ValidateOrderItems(storeID string, items []OrderItem) (*OrderValidation, error)
}

// External DTOs (for Order Service integration)
type OrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type OrderValidation struct {
	Valid       bool                 `json:"valid"`
	Items       []ValidatedOrderItem `json:"items"`
	TotalAmount float64              `json:"total_amount"`
	Errors      []string             `json:"errors,omitempty"`
}

type ValidatedOrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Available bool    `json:"available"`
	Subtotal  float64 `json:"subtotal"`
}

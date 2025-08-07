package domain

import (
	"time"

	"glovo-backend/shared/auth"
)

// Order represents the domain entity
type Order struct {
	ID                 string       `json:"id" gorm:"primaryKey"`
	CustomerID         string       `json:"customer_id" gorm:"index"`
	MerchantID         string       `json:"merchant_id" gorm:"index"`
	DriverID           *string      `json:"driver_id,omitempty" gorm:"index"`
	Status             OrderStatus  `json:"status"`
	Items              []OrderItem  `json:"items" gorm:"foreignKey:OrderID"`
	DeliveryInfo       DeliveryInfo `json:"delivery_info" gorm:"embedded"`
	PaymentInfo        PaymentInfo  `json:"payment_info" gorm:"embedded"`
	TotalAmount        float64      `json:"total_amount"`
	DeliveryFee        float64      `json:"delivery_fee"`
	ServiceFee         float64      `json:"service_fee"`
	TaxAmount          float64      `json:"tax_amount"`
	FinalAmount        float64      `json:"final_amount"`
	PlacedAt           time.Time    `json:"placed_at"`
	ScheduledFor       *time.Time   `json:"scheduled_for,omitempty"`
	EstimatedTime      *int         `json:"estimated_time,omitempty"` // in minutes
	CompletedAt        *time.Time   `json:"completed_at,omitempty"`
	CancelledAt        *time.Time   `json:"cancelled_at,omitempty"`
	CancellationReason *string      `json:"cancellation_reason,omitempty"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
}

type OrderItem struct {
	ID        string  `json:"id" gorm:"primaryKey"`
	OrderID   string  `json:"order_id"`
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Notes     string  `json:"notes,omitempty"`
}

type DeliveryInfo struct {
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Phone     string  `json:"phone"`
	Notes     string  `json:"notes,omitempty"`
}

type PaymentInfo struct {
	Method    string `json:"method"`
	Status    string `json:"status"`
	Reference string `json:"reference,omitempty"`
}

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusPreparing OrderStatus = "preparing"
	StatusReady     OrderStatus = "ready"
	StatusAssigned  OrderStatus = "assigned"
	StatusPickedUp  OrderStatus = "picked_up"
	StatusInTransit OrderStatus = "in_transit"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

// Request/Response DTOs
type CreateOrderRequest struct {
	MerchantID   string         `json:"merchant_id" binding:"required"`
	Items        []OrderItemReq `json:"items" binding:"required,min=1"`
	DeliveryInfo DeliveryInfo   `json:"delivery_info" binding:"required"`
	PaymentInfo  PaymentInfo    `json:"payment_info" binding:"required"`
	ScheduledFor *time.Time     `json:"scheduled_for,omitempty"`
	Notes        string         `json:"notes,omitempty"`
}

type OrderItemReq struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	Notes     string `json:"notes,omitempty"`
}

type UpdateOrderStatusRequest struct {
	Status             OrderStatus `json:"status" binding:"required"`
	EstimatedTime      *int        `json:"estimated_time,omitempty"`
	CancellationReason *string     `json:"cancellation_reason,omitempty"`
}

type OrderResponse struct {
	Order        Order              `json:"order"`
	MerchantInfo *MerchantInfo      `json:"merchant_info,omitempty"`
	DriverInfo   *DriverInfo        `json:"driver_info,omitempty"`
	TrackingInfo *OrderTrackingInfo `json:"tracking_info,omitempty"`
}

type MerchantInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type DriverInfo struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Phone  string  `json:"phone"`
	Rating float64 `json:"rating"`
}

type OrderTrackingInfo struct {
	CurrentLocation  *Location      `json:"current_location,omitempty"`
	EstimatedArrival *time.Time     `json:"estimated_arrival,omitempty"`
	Steps            []TrackingStep `json:"steps"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type TrackingStep struct {
	Status      OrderStatus `json:"status"`
	Description string      `json:"description"`
	Timestamp   time.Time   `json:"timestamp"`
}

// Repository interfaces (ports)
type OrderRepository interface {
	Create(order *Order) error
	GetByID(id string) (*Order, error)
	GetByCustomerID(customerID string, limit, offset int) ([]Order, error)
	GetByMerchantID(merchantID string, limit, offset int) ([]Order, error)
	GetByDriverID(driverID string, limit, offset int) ([]Order, error)
	GetByStatus(status OrderStatus, limit, offset int) ([]Order, error)
	Update(order *Order) error
	Delete(id string) error
	List(limit, offset int) ([]Order, error)
}

// Service interfaces (ports)
type OrderService interface {
	CreateOrder(customerID string, req CreateOrderRequest) (*OrderResponse, error)
	GetOrder(orderID string, userID string, role auth.UserRole) (*OrderResponse, error)
	GetOrderHistory(userID string, role auth.UserRole, limit, offset int) ([]Order, error)
	UpdateOrderStatus(orderID string, req UpdateOrderStatusRequest, userID string, role auth.UserRole) (*OrderResponse, error)
	CancelOrder(orderID string, userID string, role auth.UserRole, reason string) (*OrderResponse, error)
	GetOrdersForMerchant(merchantID string, limit, offset int) ([]Order, error)
	GetOrdersForDriver(driverID string, limit, offset int) ([]Order, error)
	GetActiveOrders() ([]Order, error)
}

// External service interfaces
type CatalogService interface {
	GetProduct(productID string) (*Product, error)
	ValidateOrder(merchantID string, items []OrderItemReq) (*OrderValidation, error)
}

type PaymentService interface {
	ProcessPayment(orderID string, amount float64, paymentInfo PaymentInfo) (*PaymentResult, error)
}

type NotificationService interface {
	SendOrderNotification(orderID string, userID string, message string) error
}

// External DTOs
type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Available   bool    `json:"available"`
}

type OrderValidation struct {
	Valid       bool            `json:"valid"`
	Items       []ValidatedItem `json:"items"`
	TotalAmount float64         `json:"total_amount"`
	Errors      []string        `json:"errors,omitempty"`
}

type ValidatedItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Available bool    `json:"available"`
	Subtotal  float64 `json:"subtotal"`
}

type PaymentResult struct {
	Success   bool   `json:"success"`
	Reference string `json:"reference"`
	Error     string `json:"error,omitempty"`
}

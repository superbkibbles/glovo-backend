package domain

import (
	"time"

	"github.com/google/uuid"
)

// OrderStatus represents order status
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusAccepted  OrderStatus = "accepted"
	OrderStatusPreparing OrderStatus = "preparing"
	OrderStatusReady     OrderStatus = "ready"
	OrderStatusAssigned  OrderStatus = "assigned"
	OrderStatusPickedUp  OrderStatus = "picked_up"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Order represents an order
type Order struct {
	ID              uuid.UUID   `gorm:"type:uuid;primary_key"`
	CustomerID      uuid.UUID   `gorm:"type:uuid;not null"`
	RestaurantID    uuid.UUID   `gorm:"type:uuid;not null"`
	RestaurantName  string
	Status          OrderStatus `gorm:"type:varchar(50);default:pending"`
	Total           float64    `gorm:"not null"`
	DeliveryAddress string
	DeliveryLat     float64
	DeliveryLng     float64
	Notes           string
	DriverID        uuid.UUID `gorm:"type:uuid"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (Order) TableName() string { return "orders" }

// OrderItem represents an order line item
type OrderItem struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	OrderID     uuid.UUID `gorm:"type:uuid;not null"`
	MenuItemID  uuid.UUID `gorm:"type:uuid;not null"`
	MenuItemName string
	Quantity    int32   `gorm:"not null"`
	UnitPrice   float64 `gorm:"not null"`
	Options     string  `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (OrderItem) TableName() string { return "order_items" }

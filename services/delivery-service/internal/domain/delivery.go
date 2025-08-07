package domain

import (
	"time"

	"glovo-backend/shared/auth"
)

// Delivery represents a delivery assignment
type Delivery struct {
	ID                 string           `json:"id" gorm:"primaryKey"`
	OrderID            string           `json:"order_id" gorm:"uniqueIndex"`
	DriverID           *string          `json:"driver_id,omitempty" gorm:"index"`
	Status             DeliveryStatus   `json:"status"`
	AssignmentType     AssignmentType   `json:"assignment_type"`
	PickupAddress      Address          `json:"pickup_address" gorm:"embedded;embeddedPrefix:pickup_"`
	DeliveryAddress    Address          `json:"delivery_address" gorm:"embedded;embeddedPrefix:delivery_"`
	EstimatedTime      int              `json:"estimated_time"`        // in minutes
	ActualTime         *int             `json:"actual_time,omitempty"` // in minutes
	Distance           float64          `json:"distance"`              // in kilometers
	DeliveryFee        float64          `json:"delivery_fee"`
	Priority           DeliveryPriority `json:"priority"`
	Notes              string           `json:"notes,omitempty"`
	AssignedAt         *time.Time       `json:"assigned_at,omitempty"`
	PickedUpAt         *time.Time       `json:"picked_up_at,omitempty"`
	DeliveredAt        *time.Time       `json:"delivered_at,omitempty"`
	CancelledAt        *time.Time       `json:"cancelled_at,omitempty"`
	CancellationReason *string          `json:"cancellation_reason,omitempty"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

type DeliveryStatus string

const (
	StatusPending   DeliveryStatus = "pending"
	StatusAssigned  DeliveryStatus = "assigned"
	StatusAccepted  DeliveryStatus = "accepted"
	StatusRejected  DeliveryStatus = "rejected"
	StatusPickedUp  DeliveryStatus = "picked_up"
	StatusInTransit DeliveryStatus = "in_transit"
	StatusDelivered DeliveryStatus = "delivered"
	StatusCancelled DeliveryStatus = "cancelled"
	StatusFailed    DeliveryStatus = "failed"
)

type AssignmentType string

const (
	AssignmentAuto   AssignmentType = "auto"
	AssignmentManual AssignmentType = "manual"
)

type DeliveryPriority string

const (
	PriorityLow    DeliveryPriority = "low"
	PriorityNormal DeliveryPriority = "normal"
	PriorityHigh   DeliveryPriority = "high"
	PriorityUrgent DeliveryPriority = "urgent"
)

type Address struct {
	Street    string  `json:"street"`
	City      string  `json:"city"`
	State     string  `json:"state"`
	ZipCode   string  `json:"zip_code"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Notes     string  `json:"notes,omitempty"`
}

// DeliveryAssignment tracks assignment attempts
type DeliveryAssignment struct {
	ID         string              `json:"id" gorm:"primaryKey"`
	DeliveryID string              `json:"delivery_id" gorm:"index"`
	DriverID   string              `json:"driver_id" gorm:"index"`
	Status     AssignmentStatus    `json:"status"`
	Response   *AssignmentResponse `json:"response,omitempty" gorm:"embedded"`
	ExpiresAt  time.Time           `json:"expires_at"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

type AssignmentStatus string

const (
	AssignmentPending  AssignmentStatus = "pending"
	AssignmentAccepted AssignmentStatus = "accepted"
	AssignmentRejected AssignmentStatus = "rejected"
	AssignmentExpired  AssignmentStatus = "expired"
)

type AssignmentResponse struct {
	ResponseTime time.Time `json:"response_time"`
	Reason       string    `json:"reason,omitempty"`
}

// DriverPerformance tracks driver metrics
type DriverPerformance struct {
	ID                  string    `json:"id" gorm:"primaryKey"`
	DriverID            string    `json:"driver_id" gorm:"uniqueIndex"`
	TotalDeliveries     int       `json:"total_deliveries"`
	CompletedDeliveries int       `json:"completed_deliveries"`
	CancelledDeliveries int       `json:"cancelled_deliveries"`
	AverageRating       float64   `json:"average_rating"`
	TotalRatings        int       `json:"total_ratings"`
	AverageDeliveryTime float64   `json:"average_delivery_time"` // in minutes
	OnTimeDeliveryRate  float64   `json:"on_time_delivery_rate"` // percentage
	AcceptanceRate      float64   `json:"acceptance_rate"`       // percentage
	LastUpdated         time.Time `json:"last_updated"`
}

// Request/Response DTOs
type CreateDeliveryRequest struct {
	OrderID         string           `json:"order_id" binding:"required"`
	PickupAddress   Address          `json:"pickup_address" binding:"required"`
	DeliveryAddress Address          `json:"delivery_address" binding:"required"`
	EstimatedTime   int              `json:"estimated_time" binding:"required"`
	Distance        float64          `json:"distance" binding:"required"`
	DeliveryFee     float64          `json:"delivery_fee" binding:"required"`
	Priority        DeliveryPriority `json:"priority"`
	Notes           string           `json:"notes,omitempty"`
}

type AssignDriverRequest struct {
	DeliveryID string         `json:"delivery_id" binding:"required"`
	DriverID   string         `json:"driver_id" binding:"required"`
	Type       AssignmentType `json:"type" binding:"required"`
}

type UpdateDeliveryStatusRequest struct {
	Status DeliveryStatus `json:"status" binding:"required"`
	Notes  string         `json:"notes,omitempty"`
}

type DriverResponseRequest struct {
	DeliveryID string `json:"delivery_id" binding:"required"`
	Accept     bool   `json:"accept"`
	Reason     string `json:"reason,omitempty"`
}

type DeliverySearchRequest struct {
	Status   DeliveryStatus   `json:"status,omitempty"`
	DriverID string           `json:"driver_id,omitempty"`
	Priority DeliveryPriority `json:"priority,omitempty"`
	DateFrom *time.Time       `json:"date_from,omitempty"`
	DateTo   *time.Time       `json:"date_to,omitempty"`
	Limit    int              `json:"limit,omitempty"`
	Offset   int              `json:"offset,omitempty"`
}

type DeliveryResponse struct {
	Delivery     Delivery      `json:"delivery"`
	DriverInfo   *DriverInfo   `json:"driver_info,omitempty"`
	OrderInfo    *OrderInfo    `json:"order_info,omitempty"`
	TrackingInfo *TrackingInfo `json:"tracking_info,omitempty"`
}

type DriverInfo struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	Rating  float64 `json:"rating"`
	Vehicle Vehicle `json:"vehicle"`
}

type Vehicle struct {
	Type         string `json:"type"`
	Make         string `json:"make"`
	Model        string `json:"model"`
	LicensePlate string `json:"license_plate"`
}

type OrderInfo struct {
	ID           string  `json:"id"`
	CustomerName string  `json:"customer_name"`
	Items        int     `json:"items"`
	TotalAmount  float64 `json:"total_amount"`
}

type TrackingInfo struct {
	CurrentLocation  *Location  `json:"current_location,omitempty"`
	EstimatedArrival *time.Time `json:"estimated_arrival,omitempty"`
	Route            []Location `json:"route,omitempty"`
}

type Location struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
}

type AutoAssignmentRequest struct {
	DeliveryID string  `json:"delivery_id" binding:"required"`
	Latitude   float64 `json:"latitude" binding:"required"`
	Longitude  float64 `json:"longitude" binding:"required"`
	Radius     float64 `json:"radius"` // in kilometers
}

type DriverAvailability struct {
	DriverID    string  `json:"driver_id"`
	Name        string  `json:"name"`
	Distance    float64 `json:"distance"`
	Rating      float64 `json:"rating"`
	ETA         int     `json:"eta"` // in minutes
	IsOnline    bool    `json:"is_online"`
	IsAvailable bool    `json:"is_available"`
}

type DeliveryMetrics struct {
	TotalDeliveries     int     `json:"total_deliveries"`
	PendingDeliveries   int     `json:"pending_deliveries"`
	ActiveDeliveries    int     `json:"active_deliveries"`
	CompletedDeliveries int     `json:"completed_deliveries"`
	CancelledDeliveries int     `json:"cancelled_deliveries"`
	AverageDeliveryTime float64 `json:"average_delivery_time"`
	OnTimeRate          float64 `json:"on_time_rate"`
	SuccessRate         float64 `json:"success_rate"`
}

// Repository interfaces (ports)
type DeliveryRepository interface {
	Create(delivery *Delivery) error
	GetByID(id string) (*Delivery, error)
	GetByOrderID(orderID string) (*Delivery, error)
	GetByDriverID(driverID string, limit, offset int) ([]Delivery, error)
	GetByStatus(status DeliveryStatus, limit, offset int) ([]Delivery, error)
	Search(req DeliverySearchRequest) ([]Delivery, error)
	Update(delivery *Delivery) error
	Delete(id string) error
	GetPendingDeliveries() ([]Delivery, error)
	GetActiveDeliveries() ([]Delivery, error)
}

type DeliveryAssignmentRepository interface {
	Create(assignment *DeliveryAssignment) error
	GetByID(id string) (*DeliveryAssignment, error)
	GetByDeliveryID(deliveryID string) ([]DeliveryAssignment, error)
	GetByDriverID(driverID string, limit, offset int) ([]DeliveryAssignment, error)
	GetPendingForDriver(driverID string) ([]DeliveryAssignment, error)
	Update(assignment *DeliveryAssignment) error
	ExpirePendingAssignments() error
}

type DriverPerformanceRepository interface {
	Create(performance *DriverPerformance) error
	GetByDriverID(driverID string) (*DriverPerformance, error)
	Update(performance *DriverPerformance) error
	GetTopDrivers(limit int) ([]DriverPerformance, error)
	GetDriverRankings() ([]DriverPerformance, error)
}

// Service interfaces (ports)
type DeliveryService interface {
	// Delivery management
	CreateDelivery(req CreateDeliveryRequest) (*DeliveryResponse, error)
	GetDelivery(deliveryID string) (*DeliveryResponse, error)
	GetDeliveryByOrder(orderID string) (*DeliveryResponse, error)
	UpdateDeliveryStatus(deliveryID string, req UpdateDeliveryStatusRequest, userID string, role auth.UserRole) (*DeliveryResponse, error)
	CancelDelivery(deliveryID string, reason string, userID string, role auth.UserRole) error

	// Driver assignment
	AutoAssignDriver(req AutoAssignmentRequest) (*DeliveryResponse, error)
	ManualAssignDriver(req AssignDriverRequest, adminID string) (*DeliveryResponse, error)
	GetAvailableDrivers(latitude, longitude, radius float64) ([]DriverAvailability, error)

	// Driver responses
	RespondToAssignment(driverID string, req DriverResponseRequest) error
	GetDriverAssignments(driverID string, limit, offset int) ([]DeliveryAssignment, error)
	GetPendingAssignments(driverID string) ([]DeliveryAssignment, error)

	// Driver operations
	PickupOrder(deliveryID string, driverID string) (*DeliveryResponse, error)
	CompleteDelivery(deliveryID string, driverID string) (*DeliveryResponse, error)
	ReportIssue(deliveryID string, driverID string, issue string) error

	// Analytics and metrics
	GetDeliveryMetrics() (*DeliveryMetrics, error)
	GetDriverPerformance(driverID string) (*DriverPerformance, error)
	UpdateDriverPerformance(driverID string) error
	GetDriverRankings() ([]DriverPerformance, error)

	// Admin operations
	SearchDeliveries(req DeliverySearchRequest) ([]Delivery, error)
	ReassignDelivery(deliveryID string, newDriverID string, adminID string) error
	GetSystemStats() (*DeliveryMetrics, error)
}

// External service interfaces
type OrderService interface {
	GetOrder(orderID string) (*OrderInfo, error)
	UpdateOrderStatus(orderID string, status string) error
}

type DriverService interface {
	GetDriver(driverID string) (*DriverInfo, error)
	GetAvailableDrivers(latitude, longitude, radius float64) ([]DriverAvailability, error)
	IsDriverAvailable(driverID string) (bool, error)
	UpdateDriverStatus(driverID string, status string) error
}

type LocationService interface {
	GetDriverLocation(driverID string) (*Location, error)
	CreateDeliveryRoute(deliveryID, driverID string, pickup, dropoff Location) error
	GetDeliveryTracking(deliveryID string) (*TrackingInfo, error)
	CalculateETA(from, to Location) (int, error)
}

type NotificationService interface {
	SendDeliveryAssignment(driverID string, delivery *Delivery) error
	SendDeliveryUpdate(orderID string, status DeliveryStatus) error
	SendDriverNotification(driverID string, message string) error
}

type PaymentService interface {
	ProcessDeliveryPayment(deliveryID string) error
	CalculateDriverPayout(deliveryID string) (float64, error)
}

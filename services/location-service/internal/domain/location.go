package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DriverLocation represents current driver location
type DriverLocation struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DriverID  string             `json:"driver_id" bson:"driver_id"`
	Location  GeoPoint           `json:"location" bson:"location"`
	Heading   float64            `json:"heading" bson:"heading"`   // Direction in degrees (0-360)
	Speed     float64            `json:"speed" bson:"speed"`       // Speed in km/h
	Accuracy  float64            `json:"accuracy" bson:"accuracy"` // GPS accuracy in meters
	Altitude  float64            `json:"altitude" bson:"altitude"` // Altitude in meters
	Status    LocationStatus     `json:"status" bson:"status"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

// LocationHistory represents historical location data
type LocationHistory struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DriverID  string             `json:"driver_id" bson:"driver_id"`
	OrderID   *string            `json:"order_id,omitempty" bson:"order_id,omitempty"`
	Route     []GeoPoint         `json:"route" bson:"route"`
	Distance  float64            `json:"distance" bson:"distance"` // Total distance in km
	Duration  int64              `json:"duration" bson:"duration"` // Duration in seconds
	StartTime time.Time          `json:"start_time" bson:"start_time"`
	EndTime   *time.Time         `json:"end_time,omitempty" bson:"end_time,omitempty"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

// GeoPoint represents a geographical point
type GeoPoint struct {
	Type        string    `json:"type" bson:"type"`               // Always "Point" for GeoJSON
	Coordinates []float64 `json:"coordinates" bson:"coordinates"` // [longitude, latitude]
	Timestamp   time.Time `json:"timestamp" bson:"timestamp"`
}

type LocationStatus string

const (
	StatusOnline  LocationStatus = "online"
	StatusOffline LocationStatus = "offline"
	StatusMoving  LocationStatus = "moving"
	StatusStopped LocationStatus = "stopped"
)

// DeliveryRoute represents a route for an active delivery
type DeliveryRoute struct {
	ID                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrderID           string             `json:"order_id" bson:"order_id"`
	DriverID          string             `json:"driver_id" bson:"driver_id"`
	PickupLocation    GeoPoint           `json:"pickup_location" bson:"pickup_location"`
	DropoffLocation   GeoPoint           `json:"dropoff_location" bson:"dropoff_location"`
	CurrentLocation   *GeoPoint          `json:"current_location,omitempty" bson:"current_location,omitempty"`
	RoutePoints       []GeoPoint         `json:"route_points" bson:"route_points"`
	Status            RouteStatus        `json:"status" bson:"status"`
	EstimatedDistance float64            `json:"estimated_distance" bson:"estimated_distance"`               // km
	EstimatedDuration int                `json:"estimated_duration" bson:"estimated_duration"`               // minutes
	ActualDistance    float64            `json:"actual_distance" bson:"actual_distance"`                     // km
	ActualDuration    *int               `json:"actual_duration,omitempty" bson:"actual_duration,omitempty"` // minutes
	StartedAt         time.Time          `json:"started_at" bson:"started_at"`
	CompletedAt       *time.Time         `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" bson:"updated_at"`
}

type RouteStatus string

const (
	RouteStatusActive    RouteStatus = "active"
	RouteStatusCompleted RouteStatus = "completed"
	RouteStatusCancelled RouteStatus = "cancelled"
)

// Geofence represents a geographical boundary
type Geofence struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Type      GeofenceType       `json:"type" bson:"type"`
	Geometry  GeofenceGeometry   `json:"geometry" bson:"geometry"`
	Metadata  map[string]string  `json:"metadata" bson:"metadata"`
	IsActive  bool               `json:"is_active" bson:"is_active"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

type GeofenceType string

const (
	GeofenceTypeDeliveryZone GeofenceType = "delivery_zone"
	GeofenceTypeRestaurant   GeofenceType = "restaurant"
	GeofenceTypeWarehouse    GeofenceType = "warehouse"
	GeofenceTypeCity         GeofenceType = "city"
)

type GeofenceGeometry struct {
	Type        string      `json:"type" bson:"type"` // "Polygon" or "Circle"
	Coordinates interface{} `json:"coordinates" bson:"coordinates"`
	Radius      *float64    `json:"radius,omitempty" bson:"radius,omitempty"` // For circles, in meters
}

// Request/Response DTOs
type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Heading   float64 `json:"heading"`
	Speed     float64 `json:"speed"`
	Accuracy  float64 `json:"accuracy"`
	Altitude  float64 `json:"altitude"`
}

type LocationResponse struct {
	DriverID  string         `json:"driver_id"`
	Location  GeoPoint       `json:"location"`
	Heading   float64        `json:"heading"`
	Speed     float64        `json:"speed"`
	Status    LocationStatus `json:"status"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type NearbyDriversRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Radius    float64 `json:"radius" binding:"required"` // in kilometers
	Limit     int     `json:"limit"`
}

type NearbyDriver struct {
	DriverID string         `json:"driver_id"`
	Location GeoPoint       `json:"location"`
	Distance float64        `json:"distance"` // in kilometers
	Status   LocationStatus `json:"status"`
}

type RouteRequest struct {
	OrderID         string   `json:"order_id" binding:"required"`
	DriverID        string   `json:"driver_id" binding:"required"`
	PickupLocation  GeoPoint `json:"pickup_location" binding:"required"`
	DropoffLocation GeoPoint `json:"dropoff_location" binding:"required"`
}

type TrackingInfo struct {
	OrderID          string      `json:"order_id"`
	DriverLocation   *GeoPoint   `json:"driver_location,omitempty"`
	EstimatedArrival *time.Time  `json:"estimated_arrival,omitempty"`
	Distance         float64     `json:"distance"`
	Duration         int         `json:"duration"`
	Status           RouteStatus `json:"status"`
}

type LocationHistoryRequest struct {
	DriverID  string    `json:"driver_id" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
	OrderID   *string   `json:"order_id,omitempty"`
}

type GeofenceEvent struct {
	DriverID   string            `json:"driver_id"`
	GeofenceID string            `json:"geofence_id"`
	EventType  GeofenceEventType `json:"event_type"`
	Location   GeoPoint          `json:"location"`
	Timestamp  time.Time         `json:"timestamp"`
}

type GeofenceEventType string

const (
	EventTypeEnter GeofenceEventType = "enter"
	EventTypeExit  GeofenceEventType = "exit"
)

// Repository interfaces (ports)
type DriverLocationRepository interface {
	Upsert(location *DriverLocation) error
	GetByDriverID(driverID string) (*DriverLocation, error)
	GetNearbyDrivers(latitude, longitude, radius float64, limit int) ([]NearbyDriver, error)
	GetDriversInGeofence(geofence *Geofence) ([]DriverLocation, error)
	UpdateStatus(driverID string, status LocationStatus) error
	Delete(driverID string) error
}

type LocationHistoryRepository interface {
	Create(history *LocationHistory) error
	GetByDriverID(driverID string, startTime, endTime time.Time) ([]LocationHistory, error)
	GetByOrderID(orderID string) (*LocationHistory, error)
	Update(history *LocationHistory) error
	Delete(id primitive.ObjectID) error
}

type DeliveryRouteRepository interface {
	Create(route *DeliveryRoute) error
	GetByID(id primitive.ObjectID) (*DeliveryRoute, error)
	GetByOrderID(orderID string) (*DeliveryRoute, error)
	GetByDriverID(driverID string) (*DeliveryRoute, error)
	GetActiveRoutes() ([]DeliveryRoute, error)
	Update(route *DeliveryRoute) error
	Complete(orderID string) error
	Delete(id primitive.ObjectID) error
}

type GeofenceRepository interface {
	Create(geofence *Geofence) error
	GetByID(id primitive.ObjectID) (*Geofence, error)
	GetAll() ([]Geofence, error)
	GetByType(geofenceType GeofenceType) ([]Geofence, error)
	Update(geofence *Geofence) error
	Delete(id primitive.ObjectID) error
}

// Service interfaces (ports)
type LocationService interface {
	// Driver location management
	UpdateDriverLocation(driverID string, req UpdateLocationRequest) (*LocationResponse, error)
	GetDriverLocation(driverID string) (*LocationResponse, error)
	GetNearbyDrivers(req NearbyDriversRequest) ([]NearbyDriver, error)
	SetDriverStatus(driverID string, status LocationStatus) error

	// Route management
	CreateDeliveryRoute(req RouteRequest) (*DeliveryRoute, error)
	GetDeliveryRoute(orderID string) (*DeliveryRoute, error)
	UpdateRouteProgress(orderID string, currentLocation GeoPoint) (*DeliveryRoute, error)
	CompleteRoute(orderID string) (*DeliveryRoute, error)
	GetActiveRoutes() ([]DeliveryRoute, error)

	// Tracking for customers
	GetOrderTrackingInfo(orderID string) (*TrackingInfo, error)

	// Location history
	GetLocationHistory(req LocationHistoryRequest) ([]LocationHistory, error)
	StartLocationTracking(driverID string, orderID *string) error
	StopLocationTracking(driverID string) error

	// Geofencing
	CreateGeofence(geofence *Geofence) (*Geofence, error)
	GetGeofences() ([]Geofence, error)
	CheckGeofenceEvents(driverID string, location GeoPoint) ([]GeofenceEvent, error)

	// Analytics
	GetDriverDistanceStats(driverID string, startTime, endTime time.Time) (*DriverStats, error)
	GetSystemLocationStats() (*SystemStats, error)
}

// External service interfaces
type MapsService interface {
	CalculateRoute(origin, destination GeoPoint) (*RouteInfo, error)
	GetETA(origin, destination GeoPoint) (*ETAInfo, error)
	ReverseGeocode(location GeoPoint) (*AddressInfo, error)
}

type NotificationService interface {
	SendGeofenceAlert(event GeofenceEvent) error
	SendLocationAlert(driverID string, message string) error
}

// External DTOs
type RouteInfo struct {
	Distance    float64    `json:"distance"` // in kilometers
	Duration    int        `json:"duration"` // in minutes
	RoutePoints []GeoPoint `json:"route_points"`
}

type ETAInfo struct {
	Duration    int       `json:"duration"` // in minutes
	Distance    float64   `json:"distance"` // in kilometers
	ArrivalTime time.Time `json:"arrival_time"`
}

type AddressInfo struct {
	Address    string `json:"address"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type DriverStats struct {
	DriverID      string  `json:"driver_id"`
	TotalDistance float64 `json:"total_distance"` // in kilometers
	TotalDuration int64   `json:"total_duration"` // in seconds
	AverageSpeed  float64 `json:"average_speed"`  // in km/h
	ActiveTime    int64   `json:"active_time"`    // in seconds
	DeliveryCount int     `json:"delivery_count"`
}

type SystemStats struct {
	ActiveDrivers       int     `json:"active_drivers"`
	TotalDistance       float64 `json:"total_distance"`
	ActiveDeliveries    int     `json:"active_deliveries"`
	AverageDeliveryTime int     `json:"average_delivery_time"`
}

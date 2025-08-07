package domain

import (
	"time"
)

// Driver represents the domain entity
type Driver struct {
	ID           string           `json:"id" gorm:"primaryKey"`
	UserID       string           `json:"user_id" gorm:"uniqueIndex"` // Links to User Service
	Status       DriverStatus     `json:"status"`
	Profile      DriverProfile    `json:"profile" gorm:"embedded"`
	Vehicle      VehicleInfo      `json:"vehicle" gorm:"embedded"`
	Documents    []DriverDocument `json:"documents" gorm:"foreignKey:DriverID"`
	Performance  PerformanceStats `json:"performance" gorm:"embedded"`
	Location     *CurrentLocation `json:"location,omitempty" gorm:"embedded"`
	Availability AvailabilityInfo `json:"availability" gorm:"embedded"`
	BankInfo     BankInfo         `json:"bank_info" gorm:"embedded"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

type DriverStatus string

const (
	StatusOffline     DriverStatus = "offline"
	StatusOnline      DriverStatus = "online"
	StatusBusy        DriverStatus = "busy"
	StatusUnavailable DriverStatus = "unavailable"
)

type DriverProfile struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Avatar      string `json:"avatar,omitempty"`
	DateOfBirth string `json:"date_of_birth"`
	Address     string `json:"address"`
}

type VehicleInfo struct {
	Type         VehicleType `json:"type"`
	Make         string      `json:"make"`
	Model        string      `json:"model"`
	Year         int         `json:"year"`
	Color        string      `json:"color"`
	LicensePlate string      `json:"license_plate"`
}

type VehicleType string

const (
	VehicleBicycle    VehicleType = "bicycle"
	VehicleMotorcycle VehicleType = "motorcycle"
	VehicleCar        VehicleType = "car"
	VehicleVan        VehicleType = "van"
)

type DriverDocument struct {
	ID         string         `json:"id" gorm:"primaryKey"`
	DriverID   string         `json:"driver_id"`
	Type       DocumentType   `json:"type"`
	URL        string         `json:"url"`
	Status     DocumentStatus `json:"status"`
	ExpiryDate *time.Time     `json:"expiry_date,omitempty"`
	UploadedAt time.Time      `json:"uploaded_at"`
}

type DocumentType string

const (
	DocDriverLicense       DocumentType = "driver_license"
	DocVehicleRegistration DocumentType = "vehicle_registration"
	DocInsurance           DocumentType = "insurance"
	DocIdentity            DocumentType = "identity"
	DocBackground          DocumentType = "background_check"
)

type DocumentStatus string

const (
	DocStatusPending  DocumentStatus = "pending"
	DocStatusApproved DocumentStatus = "approved"
	DocStatusRejected DocumentStatus = "rejected"
	DocStatusExpired  DocumentStatus = "expired"
)

type PerformanceStats struct {
	Rating              float64 `json:"rating"`
	TotalDeliveries     int     `json:"total_deliveries"`
	CompletedDeliveries int     `json:"completed_deliveries"`
	CancelledDeliveries int     `json:"cancelled_deliveries"`
	AverageDeliveryTime int     `json:"average_delivery_time"` // in minutes
	TotalEarnings       float64 `json:"total_earnings"`
	WeeklyEarnings      float64 `json:"weekly_earnings"`
	MonthlyEarnings     float64 `json:"monthly_earnings"`
	OnTimeDeliveries    int     `json:"on_time_deliveries"`
	LateDeliveries      int     `json:"late_deliveries"`
}

type CurrentLocation struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AvailabilityInfo struct {
	IsAvailable   bool          `json:"is_available"`
	WorkingHours  []WorkingHour `json:"working_hours" gorm:"serializer:json"`
	MaxDeliveries int           `json:"max_deliveries"`
	DeliveryZones []string      `json:"delivery_zones" gorm:"serializer:json"`
}

type WorkingHour struct {
	Day       string `json:"day"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type BankInfo struct {
	AccountHolder string `json:"account_holder"`
	BankName      string `json:"bank_name"`
	AccountNumber string `json:"account_number"`
	RoutingNumber string `json:"routing_number"`
}

// Request/Response DTOs
type RegisterDriverRequest struct {
	Profile  DriverProfile `json:"profile" binding:"required"`
	Vehicle  VehicleInfo   `json:"vehicle" binding:"required"`
	BankInfo BankInfo      `json:"bank_info" binding:"required"`
}

type UpdateDriverProfileRequest struct {
	Profile      *DriverProfile    `json:"profile,omitempty"`
	Vehicle      *VehicleInfo      `json:"vehicle,omitempty"`
	Availability *AvailabilityInfo `json:"availability,omitempty"`
	BankInfo     *BankInfo         `json:"bank_info,omitempty"`
}

type UpdateStatusRequest struct {
	Status DriverStatus `json:"status" binding:"required"`
}

type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

type UploadDocumentRequest struct {
	Type       DocumentType `json:"type" binding:"required"`
	URL        string       `json:"url" binding:"required"`
	ExpiryDate *time.Time   `json:"expiry_date,omitempty"`
}

type DriverSearchRequest struct {
	Status      DriverStatus `json:"status,omitempty"`
	VehicleType VehicleType  `json:"vehicle_type,omitempty"`
	Latitude    float64      `json:"latitude,omitempty"`
	Longitude   float64      `json:"longitude,omitempty"`
	Radius      float64      `json:"radius,omitempty"` // in kilometers
	MinRating   float64      `json:"min_rating,omitempty"`
	Available   *bool        `json:"available,omitempty"`
	Limit       int          `json:"limit,omitempty"`
	Offset      int          `json:"offset,omitempty"`
}

type EarningsReportRequest struct {
	DriverID  string    `json:"driver_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	GroupBy   string    `json:"group_by"` // day, week, month
}

type EarningsReport struct {
	Period      string  `json:"period"`
	Deliveries  int     `json:"deliveries"`
	Earnings    float64 `json:"earnings"`
	Commission  float64 `json:"commission"`
	NetEarnings float64 `json:"net_earnings"`
}

// Repository interfaces (ports)
type DriverRepository interface {
	Create(driver *Driver) error
	GetByID(id string) (*Driver, error)
	GetByUserID(userID string) (*Driver, error)
	Update(driver *Driver) error
	Delete(id string) error
	Search(req DriverSearchRequest) ([]Driver, error)
	List(limit, offset int) ([]Driver, error)
	GetByStatus(status DriverStatus, limit, offset int) ([]Driver, error)
}

type DriverDocumentRepository interface {
	Create(document *DriverDocument) error
	GetByID(id string) (*DriverDocument, error)
	GetByDriverID(driverID string) ([]DriverDocument, error)
	Update(document *DriverDocument) error
	Delete(id string) error
	GetByStatusAndType(status DocumentStatus, docType DocumentType) ([]DriverDocument, error)
}

// Service interfaces (ports)
type DriverService interface {
	RegisterDriver(userID string, req RegisterDriverRequest) (*Driver, error)
	GetDriver(driverID string) (*Driver, error)
	GetDriverByUser(userID string) (*Driver, error)
	UpdateProfile(driverID string, userID string, req UpdateDriverProfileRequest) (*Driver, error)
	UpdateStatus(driverID string, userID string, status DriverStatus) (*Driver, error)
	UpdateLocation(driverID string, userID string, req UpdateLocationRequest) (*Driver, error)
	UploadDocument(driverID string, userID string, req UploadDocumentRequest) (*DriverDocument, error)
	GetDocuments(driverID string, userID string) ([]DriverDocument, error)
	SearchDrivers(req DriverSearchRequest) ([]Driver, error)
	GetAvailableDrivers(latitude, longitude, radius float64) ([]Driver, error)
	GetEarningsReport(driverID string, userID string, req EarningsReportRequest) ([]EarningsReport, error)
	ApproveDocument(documentID string, adminID string) error
	RejectDocument(documentID string, adminID string, reason string) error
	UpdatePerformance(driverID string, stats PerformanceStats) error
}

// External service interfaces
type UserService interface {
	GetUser(userID string) (*User, error)
	ValidateDriverRole(userID string) error
}

type LocationService interface {
	UpdateDriverLocation(driverID string, latitude, longitude float64) error
	GetDriverLocation(driverID string) (*CurrentLocation, error)
}

type PaymentService interface {
	ProcessDriverPayout(driverID string, amount float64) error
	GetDriverEarnings(driverID string, startDate, endDate time.Time) (*EarningsReport, error)
}

// External DTOs
type User struct {
	ID    string `json:"id"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type DeliveryStats struct {
	TotalDeliveries int     `json:"total_deliveries"`
	TotalEarnings   float64 `json:"total_earnings"`
	AverageRating   float64 `json:"average_rating"`
}

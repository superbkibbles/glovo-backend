package domain

import (
	"time"

	"glovo-backend/shared/auth"
)

// Admin represents an admin user
type Admin struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	Email       string     `json:"email" gorm:"uniqueIndex"`
	Password    string     `json:"-" gorm:"not null"` // Hide password in JSON
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Role        AdminRole  `json:"role"`
	Permissions []string   `json:"permissions" gorm:"serializer:json"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type AdminRole string

const (
	RoleSuperAdmin AdminRole = "super_admin"
	RoleAdmin      AdminRole = "admin"
	RoleModerator  AdminRole = "moderator"
	RoleSupport    AdminRole = "support"
)

// Platform stats and analytics
type PlatformStats struct {
	TotalUsers        int            `json:"total_users"`
	TotalMerchants    int            `json:"total_merchants"`
	TotalDrivers      int            `json:"total_drivers"`
	ActiveOrders      int            `json:"active_orders"`
	TotalOrders       int            `json:"total_orders"`
	TotalRevenue      float64        `json:"total_revenue"`
	MonthlyRevenue    float64        `json:"monthly_revenue"`
	DailyRevenue      float64        `json:"daily_revenue"`
	AverageOrderValue float64        `json:"average_order_value"`
	TopMerchants      []MerchantStat `json:"top_merchants"`
	TopDrivers        []DriverStat   `json:"top_drivers"`
	OrdersByStatus    map[string]int `json:"orders_by_status"`
	RevenueByPeriod   []RevenueStat  `json:"revenue_by_period"`
}

type MerchantStat struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Orders  int     `json:"orders"`
	Revenue float64 `json:"revenue"`
	Rating  float64 `json:"rating"`
}

type DriverStat struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Deliveries int     `json:"deliveries"`
	Rating     float64 `json:"rating"`
	Earnings   float64 `json:"earnings"`
}

type RevenueStat struct {
	Period  string  `json:"period"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

// User management DTOs
type UserInfo struct {
	ID        string        `json:"id"`
	Phone     string        `json:"phone"`
	Email     string        `json:"email,omitempty"`
	Role      auth.UserRole `json:"role"`
	Profile   UserProfile   `json:"profile"`
	Status    UserStatus    `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type UserProfile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Avatar    string `json:"avatar,omitempty"`
}

type MerchantProfile struct {
	BusinessName string `json:"business_name"`
	Address      string `json:"Address"`
	Phone        string `json:"Phone,omitempty"`
}

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusBanned    UserStatus = "banned"
	UserStatusPending   UserStatus = "pending"
)

const (
	MerchantStatusActive    UserStatus = "active"
	MerchantStatusSuspended UserStatus = "suspended"
	MerchantStatusBanned    UserStatus = "banned"
	MerchantStatusPending   UserStatus = "pending"
)

const (
	DriverStatusActive    UserStatus = "active"
	DriverStatusSuspended UserStatus = "suspended"
	DriverStatusBanned    UserStatus = "banned"
	DriverStatusPending   UserStatus = "pending"
)

type MerchantInfo struct {
	ID           string          `json:"id"`
	UserID       string          `json:"user_id"`
	BusinessName string          `json:"business_name"`
	ContactInfo  ContactInfo     `json:"contact_info"`
	StoreInfo    StoreInfo       `json:"store_info"`
	Profile      MerchantProfile `json:"profile"`
	Status       UserStatus      `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
}

type ContactInfo struct {
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type StoreInfo struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Rating      float64 `json:"rating"`
	OrderCount  int     `json:"order_count"`
}

type DriverInfo struct {
	ID          string          `json:"id"`
	UserID      string          `json:"user_id"`
	Profile     DriverProfile   `json:"profile"`
	Vehicle     VehicleInfo     `json:"vehicle"`
	Status      UserStatus      `json:"status"`
	Performance PerformanceInfo `json:"performance"`
	CreatedAt   time.Time       `json:"created_at"`
}

type DriverProfile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

type VehicleInfo struct {
	Type         string `json:"type"`
	Make         string `json:"make"`
	Model        string `json:"model"`
	LicensePlate string `json:"license_plate"`
}

type PerformanceInfo struct {
	Rating          float64 `json:"rating"`
	TotalDeliveries int     `json:"total_deliveries"`
	TotalEarnings   float64 `json:"total_earnings"`
}

// Request/Response DTOs
type AdminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AdminLoginResponse struct {
	Token     string `json:"token"`
	Admin     Admin  `json:"admin"`
	ExpiresAt int64  `json:"expires_at"`
}

type CreateAdminRequest struct {
	Email       string    `json:"email" binding:"required,email"`
	Password    string    `json:"password" binding:"required,min=8"`
	FirstName   string    `json:"first_name" binding:"required"`
	LastName    string    `json:"last_name" binding:"required"`
	Role        AdminRole `json:"role" binding:"required"`
	Permissions []string  `json:"permissions"`
}

type UpdateAdminRequest struct {
	FirstName   *string    `json:"first_name,omitempty"`
	LastName    *string    `json:"last_name,omitempty"`
	Role        *AdminRole `json:"role,omitempty"`
	Permissions []string   `json:"permissions,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
}

type UserSearchRequest struct {
	Query  string        `json:"query,omitempty"`
	Role   auth.UserRole `json:"role,omitempty"`
	Status UserStatus    `json:"status,omitempty"`
	Limit  int           `json:"limit,omitempty"`
	Offset int           `json:"offset,omitempty"`
}

type UpdateUserStatusRequest struct {
	Status UserStatus `json:"status" binding:"required"`
	Reason string     `json:"reason,omitempty"`
}

type SystemConfigRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type SystemConfig struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Key         string    `json:"key" gorm:"uniqueIndex"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
	UpdatedBy   string    `json:"updated_by"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type AuditLog struct {
	ID         string                 `json:"id" gorm:"primaryKey"`
	AdminID    string                 `json:"admin_id" gorm:"index"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id"`
	Details    map[string]interface{} `json:"details" gorm:"serializer:json"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	CreatedAt  time.Time              `json:"created_at"`
}

// Repository interfaces (ports)
type AdminRepository interface {
	Create(admin *Admin) error
	GetByID(id string) (*Admin, error)
	GetByEmail(email string) (*Admin, error)
	Update(admin *Admin) error
	Delete(id string) error
	List(limit, offset int) ([]Admin, error)
	UpdateLastLogin(id string) error
}

type SystemConfigRepository interface {
	Set(config *SystemConfig) error
	Get(key string) (*SystemConfig, error)
	GetAll() ([]SystemConfig, error)
	Delete(key string) error
}

type AuditLogRepository interface {
	Create(log *AuditLog) error
	GetByAdminID(adminID string, limit, offset int) ([]AuditLog, error)
	GetByResource(resource string, limit, offset int) ([]AuditLog, error)
	List(limit, offset int) ([]AuditLog, error)
}

// Service interfaces (ports)
type AdminService interface {
	// Admin authentication
	Login(req AdminLoginRequest) (*AdminLoginResponse, error)
	GetProfile(adminID string) (*Admin, error)

	// Admin management
	CreateAdmin(adminID string, req CreateAdminRequest) (*Admin, error)
	UpdateAdmin(adminID, targetAdminID string, req UpdateAdminRequest) (*Admin, error)
	DeleteAdmin(adminID, targetAdminID string) error
	ListAdmins(adminID string, limit, offset int) ([]Admin, error)

	// User management
	GetUsers(req UserSearchRequest) ([]UserInfo, error)
	GetUser(userID string) (*UserInfo, error)
	UpdateUserStatus(adminID, userID string, req UpdateUserStatusRequest) error
	GetMerchants(limit, offset int) ([]MerchantInfo, error)
	GetDrivers(limit, offset int) ([]DriverInfo, error)
	SuspendUser(adminID, userID string, reason string) error
	ReactivateUser(adminID, userID string) error

	// Platform analytics
	GetPlatformStats() (*PlatformStats, error)
	GetRevenueStats(startDate, endDate time.Time) ([]RevenueStat, error)
	GetOrderStats(startDate, endDate time.Time) (map[string]int, error)

	// System configuration
	GetSystemConfig(key string) (*SystemConfig, error)
	SetSystemConfig(adminID string, req SystemConfigRequest) (*SystemConfig, error)
	GetAllSystemConfigs() ([]SystemConfig, error)

	// Audit logging
	LogAction(adminID, action, resource, resourceID string, details map[string]interface{}, ipAddress, userAgent string) error
	GetAuditLogs(adminID string, limit, offset int) ([]AuditLog, error)
	GetResourceAuditLogs(resource string, limit, offset int) ([]AuditLog, error)
}

// External service interfaces
type UserService interface {
	GetUser(userID string) (*UserInfo, error)
	GetUsers(req UserSearchRequest) ([]UserInfo, error)
	UpdateUserStatus(userID string, status UserStatus, reason string) error
	GetMerchants(limit, offset int) ([]MerchantInfo, error)
	GetDrivers(limit, offset int) ([]DriverInfo, error)
}

type OrderService interface {
	GetOrderStats(startDate, endDate time.Time) (map[string]int, error)
	GetActiveOrdersCount() (int, error)
	GetTotalOrdersCount() (int, error)
}

type PaymentService interface {
	GetRevenueStats(startDate, endDate time.Time) ([]RevenueStat, error)
	GetTotalRevenue() (float64, error)
	GetAverageOrderValue() (float64, error)
}

type CatalogService interface {
	GetTopMerchants(limit int) ([]MerchantStat, error)
}

type DriverService interface {
	GetTopDrivers(limit int) ([]DriverStat, error)
}

type AnalyticsService interface {
	GetPlatformStats() (*PlatformStats, error)
	GetRevenueAnalytics(startDate, endDate time.Time) ([]RevenueStat, error)
}

package domain

import (
	"time"
)

// Analytics entities
type PlatformMetrics struct {
	ID                string    `json:"id" gorm:"primaryKey"`
	Date              time.Time `json:"date" gorm:"index"`
	TotalUsers        int       `json:"total_users"`
	TotalMerchants    int       `json:"total_merchants"`
	TotalDrivers      int       `json:"total_drivers"`
	ActiveOrders      int       `json:"active_orders"`
	CompletedOrders   int       `json:"completed_orders"`
	CancelledOrders   int       `json:"cancelled_orders"`
	TotalRevenue      float64   `json:"total_revenue"`
	DailyRevenue      float64   `json:"daily_revenue"`
	AverageOrderValue float64   `json:"average_order_value"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type RevenueMetrics struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	Date            time.Time `json:"date" gorm:"index"`
	TotalRevenue    float64   `json:"total_revenue"`
	OrderCommission float64   `json:"order_commission"`
	DeliveryFees    float64   `json:"delivery_fees"`
	DriverPayouts   float64   `json:"driver_payouts"`
	MerchantPayouts float64   `json:"merchant_payouts"`
	NetProfit       float64   `json:"net_profit"`
	OrderCount      int       `json:"order_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type DriverMetrics struct {
	ID                  string    `json:"id" gorm:"primaryKey"`
	DriverID            string    `json:"driver_id" gorm:"index"`
	Date                time.Time `json:"date" gorm:"index"`
	TotalDeliveries     int       `json:"total_deliveries"`
	CompletedDeliveries int       `json:"completed_deliveries"`
	CancelledDeliveries int       `json:"cancelled_deliveries"`
	TotalEarnings       float64   `json:"total_earnings"`
	AverageRating       float64   `json:"average_rating"`
	OnlineHours         float64   `json:"online_hours"`
	DistanceTraveled    float64   `json:"distance_traveled"`
	AcceptanceRate      float64   `json:"acceptance_rate"`
	CreatedAt           time.Time `json:"created_at"`
}

type MerchantMetrics struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	MerchantID      string    `json:"merchant_id" gorm:"index"`
	Date            time.Time `json:"date" gorm:"index"`
	TotalOrders     int       `json:"total_orders"`
	CompletedOrders int       `json:"completed_orders"`
	CancelledOrders int       `json:"cancelled_orders"`
	TotalRevenue    float64   `json:"total_revenue"`
	Commission      float64   `json:"commission"`
	NetRevenue      float64   `json:"net_revenue"`
	AverageRating   float64   `json:"average_rating"`
	CreatedAt       time.Time `json:"created_at"`
}

// Response DTOs
type PlatformStats struct {
	TotalUsers        int               `json:"total_users"`
	TotalMerchants    int               `json:"total_merchants"`
	TotalDrivers      int               `json:"total_drivers"`
	ActiveOrders      int               `json:"active_orders"`
	TotalOrders       int               `json:"total_orders"`
	TotalRevenue      float64           `json:"total_revenue"`
	MonthlyRevenue    float64           `json:"monthly_revenue"`
	DailyRevenue      float64           `json:"daily_revenue"`
	AverageOrderValue float64           `json:"average_order_value"`
	TopMerchants      []MerchantSummary `json:"top_merchants"`
	TopDrivers        []DriverSummary   `json:"top_drivers"`
	OrdersByStatus    map[string]int    `json:"orders_by_status"`
	RevenueByPeriod   []RevenueStat     `json:"revenue_by_period"`
	GrowthMetrics     GrowthMetrics     `json:"growth_metrics"`
}

type MerchantSummary struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Orders  int     `json:"orders"`
	Revenue float64 `json:"revenue"`
	Rating  float64 `json:"rating"`
}

type DriverSummary struct {
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

type GrowthMetrics struct {
	UserGrowthRate     float64 `json:"user_growth_rate"`
	OrderGrowthRate    float64 `json:"order_growth_rate"`
	RevenueGrowthRate  float64 `json:"revenue_growth_rate"`
	MerchantGrowthRate float64 `json:"merchant_growth_rate"`
	DriverGrowthRate   float64 `json:"driver_growth_rate"`
}

type TimeSeriesRequest struct {
	StartDate time.Time `json:"start_date" binding:"required"`
	EndDate   time.Time `json:"end_date" binding:"required"`
	Interval  string    `json:"interval"` // daily, weekly, monthly
}

type AnalyticsRequest struct {
	StartDate  time.Time              `json:"start_date" binding:"required"`
	EndDate    time.Time              `json:"end_date" binding:"required"`
	MetricType string                 `json:"metric_type,omitempty"`
	GroupBy    string                 `json:"group_by,omitempty"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
}

type TopPerformersRequest struct {
	Period string `json:"period"` // today, week, month, year
	Limit  int    `json:"limit"`
	Type   string `json:"type"` // merchants, drivers
}

// Repository interfaces (ports)
type PlatformMetricsRepository interface {
	Create(metrics *PlatformMetrics) error
	GetByDate(date time.Time) (*PlatformMetrics, error)
	GetByDateRange(startDate, endDate time.Time) ([]PlatformMetrics, error)
	GetLatest() (*PlatformMetrics, error)
	Update(metrics *PlatformMetrics) error
}

type RevenueMetricsRepository interface {
	Create(metrics *RevenueMetrics) error
	GetByDate(date time.Time) (*RevenueMetrics, error)
	GetByDateRange(startDate, endDate time.Time) ([]RevenueMetrics, error)
	GetTotalRevenue() (float64, error)
	GetRevenueByPeriod(startDate, endDate time.Time, interval string) ([]RevenueStat, error)
}

type DriverMetricsRepository interface {
	Create(metrics *DriverMetrics) error
	GetByDriverID(driverID string, startDate, endDate time.Time) ([]DriverMetrics, error)
	GetTopDrivers(period string, limit int) ([]DriverSummary, error)
	GetDriverPerformance(driverID string) (*DriverMetrics, error)
	AggregateDriverMetrics(startDate, endDate time.Time) (*DriverMetrics, error)
}

type MerchantMetricsRepository interface {
	Create(metrics *MerchantMetrics) error
	GetByMerchantID(merchantID string, startDate, endDate time.Time) ([]MerchantMetrics, error)
	GetTopMerchants(period string, limit int) ([]MerchantSummary, error)
	GetMerchantPerformance(merchantID string) (*MerchantMetrics, error)
	AggregateMerchantMetrics(startDate, endDate time.Time) (*MerchantMetrics, error)
}

// Service interfaces (ports)
type AnalyticsService interface {
	// Platform analytics
	GetPlatformStats() (*PlatformStats, error)
	GetRevenueAnalytics(startDate, endDate time.Time) ([]RevenueStat, error)
	GetGrowthMetrics(period string) (*GrowthMetrics, error)
	GetTimeSeriesData(req TimeSeriesRequest) (interface{}, error)

	// Performance analytics
	GetTopMerchants(req TopPerformersRequest) ([]MerchantSummary, error)
	GetTopDrivers(req TopPerformersRequest) ([]DriverSummary, error)
	GetMerchantAnalytics(merchantID string, startDate, endDate time.Time) (*MerchantMetrics, error)
	GetDriverAnalytics(driverID string, startDate, endDate time.Time) (*DriverMetrics, error)

	// Order analytics
	GetOrderStatistics(startDate, endDate time.Time) (map[string]int, error)
	GetOrderTrends(startDate, endDate time.Time) ([]interface{}, error)

	// Real-time metrics
	GetRealTimeStats() (map[string]interface{}, error)

	// Data aggregation
	AggregateDaily() error
	RefreshMetrics() error
}

// External service interfaces
type UserService interface {
	GetUserCount() (int, error)
	GetMerchantCount() (int, error)
	GetDriverCount() (int, error)
	GetUserGrowthRate(period string) (float64, error)
}

type OrderService interface {
	GetOrderCount() (int, error)
	GetActiveOrdersCount() (int, error)
	GetCompletedOrdersCount() (int, error)
	GetCancelledOrdersCount() (int, error)
	GetOrdersByStatus() (map[string]int, error)
	GetOrderGrowthRate(period string) (float64, error)
	GetOrderTrends(startDate, endDate time.Time) ([]interface{}, error)
}

type PaymentService interface {
	GetTotalRevenue() (float64, error)
	GetDailyRevenue() (float64, error)
	GetMonthlyRevenue() (float64, error)
	GetAverageOrderValue() (float64, error)
	GetRevenueByPeriod(startDate, endDate time.Time) ([]RevenueStat, error)
	GetRevenueGrowthRate(period string) (float64, error)
}

type CatalogService interface {
	GetTopMerchants(limit int) ([]MerchantSummary, error)
	GetMerchantPerformance(merchantID string) (*MerchantMetrics, error)
}

type DriverService interface {
	GetTopDrivers(limit int) ([]DriverSummary, error)
	GetDriverPerformance(driverID string) (*DriverMetrics, error)
}

type LocationService interface {
	GetDriverMetrics(driverID string, startDate, endDate time.Time) (map[string]interface{}, error)
}

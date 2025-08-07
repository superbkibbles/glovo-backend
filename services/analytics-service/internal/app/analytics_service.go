package app

import (
	"time"

	"glovo-backend/services/analytics-service/internal/domain"

	"github.com/google/uuid"
)

type analyticsService struct {
	platformRepo    domain.PlatformMetricsRepository
	revenueRepo     domain.RevenueMetricsRepository
	driverRepo      domain.DriverMetricsRepository
	merchantRepo    domain.MerchantMetricsRepository
	userService     domain.UserService
	orderService    domain.OrderService
	paymentService  domain.PaymentService
	catalogService  domain.CatalogService
	driverService   domain.DriverService
	locationService domain.LocationService
}

func NewAnalyticsService(
	platformRepo domain.PlatformMetricsRepository,
	revenueRepo domain.RevenueMetricsRepository,
	driverRepo domain.DriverMetricsRepository,
	merchantRepo domain.MerchantMetricsRepository,
	userService domain.UserService,
	orderService domain.OrderService,
	paymentService domain.PaymentService,
	catalogService domain.CatalogService,
	driverService domain.DriverService,
	locationService domain.LocationService,
) domain.AnalyticsService {
	return &analyticsService{
		platformRepo:    platformRepo,
		revenueRepo:     revenueRepo,
		driverRepo:      driverRepo,
		merchantRepo:    merchantRepo,
		userService:     userService,
		orderService:    orderService,
		paymentService:  paymentService,
		catalogService:  catalogService,
		driverService:   driverService,
		locationService: locationService,
	}
}

// Platform analytics
func (s *analyticsService) GetPlatformStats() (*domain.PlatformStats, error) {
	stats := &domain.PlatformStats{}

	// Get user counts
	if totalUsers, err := s.userService.GetUserCount(); err == nil {
		stats.TotalUsers = totalUsers
	}

	if totalMerchants, err := s.userService.GetMerchantCount(); err == nil {
		stats.TotalMerchants = totalMerchants
	}

	if totalDrivers, err := s.userService.GetDriverCount(); err == nil {
		stats.TotalDrivers = totalDrivers
	}

	// Get order stats
	if activeOrders, err := s.orderService.GetActiveOrdersCount(); err == nil {
		stats.ActiveOrders = activeOrders
	}

	if totalOrders, err := s.orderService.GetOrderCount(); err == nil {
		stats.TotalOrders = totalOrders
	}

	if ordersByStatus, err := s.orderService.GetOrdersByStatus(); err == nil {
		stats.OrdersByStatus = ordersByStatus
	}

	// Get revenue stats
	if totalRevenue, err := s.paymentService.GetTotalRevenue(); err == nil {
		stats.TotalRevenue = totalRevenue
	}

	if dailyRevenue, err := s.paymentService.GetDailyRevenue(); err == nil {
		stats.DailyRevenue = dailyRevenue
	}

	if monthlyRevenue, err := s.paymentService.GetMonthlyRevenue(); err == nil {
		stats.MonthlyRevenue = monthlyRevenue
	}

	if avgOrderValue, err := s.paymentService.GetAverageOrderValue(); err == nil {
		stats.AverageOrderValue = avgOrderValue
	}

	// Get top performers
	if topMerchants, err := s.catalogService.GetTopMerchants(5); err == nil {
		stats.TopMerchants = topMerchants
	}

	if topDrivers, err := s.driverService.GetTopDrivers(5); err == nil {
		stats.TopDrivers = topDrivers
	}

	// Get revenue by period (last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)
	if revenuePeriod, err := s.paymentService.GetRevenueByPeriod(startDate, endDate); err == nil {
		stats.RevenueByPeriod = revenuePeriod
	}

	// Get growth metrics
	if growthMetrics, err := s.GetGrowthMetrics("month"); err == nil {
		stats.GrowthMetrics = *growthMetrics
	}

	return stats, nil
}

func (s *analyticsService) GetRevenueAnalytics(startDate, endDate time.Time) ([]domain.RevenueStat, error) {
	return s.paymentService.GetRevenueByPeriod(startDate, endDate)
}

func (s *analyticsService) GetGrowthMetrics(period string) (*domain.GrowthMetrics, error) {
	metrics := &domain.GrowthMetrics{}

	// Get growth rates from external services
	if userGrowth, err := s.userService.GetUserGrowthRate(period); err == nil {
		metrics.UserGrowthRate = userGrowth
	}

	if orderGrowth, err := s.orderService.GetOrderGrowthRate(period); err == nil {
		metrics.OrderGrowthRate = orderGrowth
	}

	if revenueGrowth, err := s.paymentService.GetRevenueGrowthRate(period); err == nil {
		metrics.RevenueGrowthRate = revenueGrowth
	}

	// Mock values for merchant and driver growth (would be calculated from user service)
	metrics.MerchantGrowthRate = 5.2
	metrics.DriverGrowthRate = 8.1

	return metrics, nil
}

func (s *analyticsService) GetTimeSeriesData(req domain.TimeSeriesRequest) (interface{}, error) {
	// Return platform metrics for the requested time range
	return s.platformRepo.GetByDateRange(req.StartDate, req.EndDate)
}

// Performance analytics
func (s *analyticsService) GetTopMerchants(req domain.TopPerformersRequest) ([]domain.MerchantSummary, error) {
	if req.Limit == 0 {
		req.Limit = 10
	}
	return s.merchantRepo.GetTopMerchants(req.Period, req.Limit)
}

func (s *analyticsService) GetTopDrivers(req domain.TopPerformersRequest) ([]domain.DriverSummary, error) {
	if req.Limit == 0 {
		req.Limit = 10
	}
	return s.driverRepo.GetTopDrivers(req.Period, req.Limit)
}

func (s *analyticsService) GetMerchantAnalytics(merchantID string, startDate, endDate time.Time) (*domain.MerchantMetrics, error) {
	// Get aggregated metrics for the merchant
	return s.merchantRepo.AggregateMerchantMetrics(startDate, endDate)
}

func (s *analyticsService) GetDriverAnalytics(driverID string, startDate, endDate time.Time) (*domain.DriverMetrics, error) {
	// Get aggregated metrics for the driver
	return s.driverRepo.AggregateDriverMetrics(startDate, endDate)
}

// Order analytics
func (s *analyticsService) GetOrderStatistics(startDate, endDate time.Time) (map[string]int, error) {
	return s.orderService.GetOrdersByStatus()
}

func (s *analyticsService) GetOrderTrends(startDate, endDate time.Time) ([]interface{}, error) {
	return s.orderService.GetOrderTrends(startDate, endDate)
}

// Real-time metrics
func (s *analyticsService) GetRealTimeStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get real-time counts
	if activeOrders, err := s.orderService.GetActiveOrdersCount(); err == nil {
		stats["active_orders"] = activeOrders
	}

	if dailyRevenue, err := s.paymentService.GetDailyRevenue(); err == nil {
		stats["daily_revenue"] = dailyRevenue
	}

	// Add timestamp
	stats["timestamp"] = time.Now()
	stats["status"] = "live"

	return stats, nil
}

// Data aggregation
func (s *analyticsService) AggregateDaily() error {
	today := time.Now().Truncate(24 * time.Hour)

	// Check if metrics already exist for today
	if existing, _ := s.platformRepo.GetByDate(today); existing != nil {
		return s.updateDailyMetrics(existing)
	}

	// Create new daily metrics
	return s.createDailyMetrics(today)
}

func (s *analyticsService) RefreshMetrics() error {
	// Refresh all cached metrics
	return s.AggregateDaily()
}

// Helper methods
func (s *analyticsService) createDailyMetrics(date time.Time) error {
	metrics := &domain.PlatformMetrics{
		ID:        uuid.New().String(),
		Date:      date,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Populate metrics from external services
	if totalUsers, err := s.userService.GetUserCount(); err == nil {
		metrics.TotalUsers = totalUsers
	}

	if totalMerchants, err := s.userService.GetMerchantCount(); err == nil {
		metrics.TotalMerchants = totalMerchants
	}

	if totalDrivers, err := s.userService.GetDriverCount(); err == nil {
		metrics.TotalDrivers = totalDrivers
	}

	if activeOrders, err := s.orderService.GetActiveOrdersCount(); err == nil {
		metrics.ActiveOrders = activeOrders
	}

	if completedOrders, err := s.orderService.GetCompletedOrdersCount(); err == nil {
		metrics.CompletedOrders = completedOrders
	}

	if cancelledOrders, err := s.orderService.GetCancelledOrdersCount(); err == nil {
		metrics.CancelledOrders = cancelledOrders
	}

	if totalRevenue, err := s.paymentService.GetTotalRevenue(); err == nil {
		metrics.TotalRevenue = totalRevenue
	}

	if dailyRevenue, err := s.paymentService.GetDailyRevenue(); err == nil {
		metrics.DailyRevenue = dailyRevenue
	}

	if avgOrderValue, err := s.paymentService.GetAverageOrderValue(); err == nil {
		metrics.AverageOrderValue = avgOrderValue
	}

	return s.platformRepo.Create(metrics)
}

func (s *analyticsService) updateDailyMetrics(metrics *domain.PlatformMetrics) error {
	// Update existing metrics with current data
	if totalUsers, err := s.userService.GetUserCount(); err == nil {
		metrics.TotalUsers = totalUsers
	}

	if activeOrders, err := s.orderService.GetActiveOrdersCount(); err == nil {
		metrics.ActiveOrders = activeOrders
	}

	if dailyRevenue, err := s.paymentService.GetDailyRevenue(); err == nil {
		metrics.DailyRevenue = dailyRevenue
	}

	metrics.UpdatedAt = time.Now()
	return s.platformRepo.Update(metrics)
}

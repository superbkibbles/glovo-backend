package app

import (
	"errors"
	"fmt"
	"time"

	"glovo-backend/services/admin-service/internal/domain"
	"glovo-backend/shared/auth"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type adminService struct {
	adminRepo        domain.AdminRepository
	systemConfigRepo domain.SystemConfigRepository
	auditLogRepo     domain.AuditLogRepository
	userService      domain.UserService
	orderService     domain.OrderService
	paymentService   domain.PaymentService
	catalogService   domain.CatalogService
	driverService    domain.DriverService
	analyticsService domain.AnalyticsService
}

func NewAdminService(
	adminRepo domain.AdminRepository,
	systemConfigRepo domain.SystemConfigRepository,
	auditLogRepo domain.AuditLogRepository,
	userService domain.UserService,
	orderService domain.OrderService,
	paymentService domain.PaymentService,
	catalogService domain.CatalogService,
	driverService domain.DriverService,
	analyticsService domain.AnalyticsService,
) domain.AdminService {
	return &adminService{
		adminRepo:        adminRepo,
		systemConfigRepo: systemConfigRepo,
		auditLogRepo:     auditLogRepo,
		userService:      userService,
		orderService:     orderService,
		paymentService:   paymentService,
		catalogService:   catalogService,
		driverService:    driverService,
		analyticsService: analyticsService,
	}
}

// Admin authentication
func (s *adminService) Login(req domain.AdminLoginRequest) (*domain.AdminLoginResponse, error) {
	admin, err := s.adminRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !admin.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(admin.ID, auth.RoleAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Update last login
	s.adminRepo.UpdateLastLogin(admin.ID)

	// Log the login action
	s.LogAction(admin.ID, "login", "admin", admin.ID, map[string]interface{}{
		"email": admin.Email,
	}, "", "")

	return &domain.AdminLoginResponse{
		Token:     token,
		Admin:     *admin,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}, nil
}

func (s *adminService) GetProfile(adminID string) (*domain.Admin, error) {
	return s.adminRepo.GetByID(adminID)
}

// Admin management
func (s *adminService) CreateAdmin(adminID string, req domain.CreateAdminRequest) (*domain.Admin, error) {
	// Check if requesting admin has permission
	requestingAdmin, err := s.adminRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}

	if requestingAdmin.Role != domain.RoleSuperAdmin {
		return nil, errors.New("insufficient permissions to create admin")
	}

	// Check if email already exists
	if _, err := s.adminRepo.GetByEmail(req.Email); err == nil {
		return nil, errors.New("admin with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	admin := &domain.Admin{
		ID:          uuid.New().String(),
		Email:       req.Email,
		Password:    string(hashedPassword),
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Role:        req.Role,
		Permissions: req.Permissions,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.adminRepo.Create(admin); err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	// Log the action
	s.LogAction(adminID, "create_admin", "admin", admin.ID, map[string]interface{}{
		"email": admin.Email,
		"role":  admin.Role,
	}, "", "")

	// Clear password before returning
	admin.Password = ""
	return admin, nil
}

func (s *adminService) UpdateAdmin(adminID, targetAdminID string, req domain.UpdateAdminRequest) (*domain.Admin, error) {
	requestingAdmin, err := s.adminRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}

	targetAdmin, err := s.adminRepo.GetByID(targetAdminID)
	if err != nil {
		return nil, err
	}

	// Permission check
	if requestingAdmin.Role != domain.RoleSuperAdmin && adminID != targetAdminID {
		return nil, errors.New("insufficient permissions")
	}

	// Apply updates
	if req.FirstName != nil {
		targetAdmin.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		targetAdmin.LastName = *req.LastName
	}
	if req.Role != nil && requestingAdmin.Role == domain.RoleSuperAdmin {
		targetAdmin.Role = *req.Role
	}
	if req.Permissions != nil && requestingAdmin.Role == domain.RoleSuperAdmin {
		targetAdmin.Permissions = req.Permissions
	}
	if req.IsActive != nil && requestingAdmin.Role == domain.RoleSuperAdmin {
		targetAdmin.IsActive = *req.IsActive
	}

	targetAdmin.UpdatedAt = time.Now()

	if err := s.adminRepo.Update(targetAdmin); err != nil {
		return nil, fmt.Errorf("failed to update admin: %w", err)
	}

	// Log the action
	s.LogAction(adminID, "update_admin", "admin", targetAdminID, map[string]interface{}{
		"changes": req,
	}, "", "")

	targetAdmin.Password = ""
	return targetAdmin, nil
}

func (s *adminService) DeleteAdmin(adminID, targetAdminID string) error {
	requestingAdmin, err := s.adminRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	if requestingAdmin.Role != domain.RoleSuperAdmin {
		return errors.New("insufficient permissions")
	}

	if adminID == targetAdminID {
		return errors.New("cannot delete yourself")
	}

	if err := s.adminRepo.Delete(targetAdminID); err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}

	// Log the action
	s.LogAction(adminID, "delete_admin", "admin", targetAdminID, nil, "", "")

	return nil
}

func (s *adminService) ListAdmins(adminID string, limit, offset int) ([]domain.Admin, error) {
	requestingAdmin, err := s.adminRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}

	if requestingAdmin.Role != domain.RoleSuperAdmin && requestingAdmin.Role != domain.RoleAdmin {
		return nil, errors.New("insufficient permissions")
	}

	admins, err := s.adminRepo.List(limit, offset)
	if err != nil {
		return nil, err
	}

	// Clear passwords
	for i := range admins {
		admins[i].Password = ""
	}

	return admins, nil
}

// User management
func (s *adminService) GetUsers(req domain.UserSearchRequest) ([]domain.UserInfo, error) {
	return s.userService.GetUsers(req)
}

func (s *adminService) GetUser(userID string) (*domain.UserInfo, error) {
	return s.userService.GetUser(userID)
}

func (s *adminService) UpdateUserStatus(adminID, userID string, req domain.UpdateUserStatusRequest) error {
	if err := s.userService.UpdateUserStatus(userID, req.Status, req.Reason); err != nil {
		return err
	}

	// Log the action
	s.LogAction(adminID, "update_user_status", "user", userID, map[string]interface{}{
		"status": req.Status,
		"reason": req.Reason,
	}, "", "")

	return nil
}

func (s *adminService) GetMerchants(limit, offset int) ([]domain.MerchantInfo, error) {
	return s.userService.GetMerchants(limit, offset)
}

func (s *adminService) GetDrivers(limit, offset int) ([]domain.DriverInfo, error) {
	return s.userService.GetDrivers(limit, offset)
}

func (s *adminService) SuspendUser(adminID, userID string, reason string) error {
	req := domain.UpdateUserStatusRequest{
		Status: domain.UserStatusSuspended,
		Reason: reason,
	}
	return s.UpdateUserStatus(adminID, userID, req)
}

func (s *adminService) ReactivateUser(adminID, userID string) error {
	req := domain.UpdateUserStatusRequest{
		Status: domain.UserStatusActive,
		Reason: "Reactivated by admin",
	}
	return s.UpdateUserStatus(adminID, userID, req)
}

// Platform analytics
func (s *adminService) GetPlatformStats() (*domain.PlatformStats, error) {
	// If analytics service is available, use it
	if s.analyticsService != nil {
		return s.analyticsService.GetPlatformStats()
	}

	// Otherwise, aggregate from individual services
	stats := &domain.PlatformStats{}

	// Get user counts
	users, err := s.userService.GetUsers(domain.UserSearchRequest{Limit: 1})
	if err == nil {
		stats.TotalUsers = len(users) // This is a simplified count
	}

	// Get merchant and driver counts
	merchants, err := s.GetMerchants(1000, 0)
	if err == nil {
		stats.TotalMerchants = len(merchants)
	}

	drivers, err := s.GetDrivers(1000, 0)
	if err == nil {
		stats.TotalDrivers = len(drivers)
	}

	// Get order stats
	if s.orderService != nil {
		if activeOrders, err := s.orderService.GetActiveOrdersCount(); err == nil {
			stats.ActiveOrders = activeOrders
		}
		if totalOrders, err := s.orderService.GetTotalOrdersCount(); err == nil {
			stats.TotalOrders = totalOrders
		}
	}

	// Get revenue stats
	if s.paymentService != nil {
		if totalRevenue, err := s.paymentService.GetTotalRevenue(); err == nil {
			stats.TotalRevenue = totalRevenue
		}
		if avgOrderValue, err := s.paymentService.GetAverageOrderValue(); err == nil {
			stats.AverageOrderValue = avgOrderValue
		}
	}

	// Get top performers
	if s.catalogService != nil {
		if topMerchants, err := s.catalogService.GetTopMerchants(5); err == nil {
			stats.TopMerchants = topMerchants
		}
	}

	if s.driverService != nil {
		if topDrivers, err := s.driverService.GetTopDrivers(5); err == nil {
			stats.TopDrivers = topDrivers
		}
	}

	return stats, nil
}

func (s *adminService) GetRevenueStats(startDate, endDate time.Time) ([]domain.RevenueStat, error) {
	if s.analyticsService != nil {
		return s.analyticsService.GetRevenueAnalytics(startDate, endDate)
	}

	if s.paymentService != nil {
		return s.paymentService.GetRevenueStats(startDate, endDate)
	}

	return nil, errors.New("analytics service not available")
}

func (s *adminService) GetOrderStats(startDate, endDate time.Time) (map[string]int, error) {
	if s.orderService != nil {
		return s.orderService.GetOrderStats(startDate, endDate)
	}

	return nil, errors.New("order service not available")
}

// System configuration
func (s *adminService) GetSystemConfig(key string) (*domain.SystemConfig, error) {
	return s.systemConfigRepo.Get(key)
}

func (s *adminService) SetSystemConfig(adminID string, req domain.SystemConfigRequest) (*domain.SystemConfig, error) {
	config := &domain.SystemConfig{
		ID:        uuid.New().String(),
		Key:       req.Key,
		Value:     req.Value,
		UpdatedBy: adminID,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	if err := s.systemConfigRepo.Set(config); err != nil {
		return nil, fmt.Errorf("failed to set system config: %w", err)
	}

	// Log the action
	s.LogAction(adminID, "set_system_config", "system_config", req.Key, map[string]interface{}{
		"key":   req.Key,
		"value": req.Value,
	}, "", "")

	return config, nil
}

func (s *adminService) GetAllSystemConfigs() ([]domain.SystemConfig, error) {
	return s.systemConfigRepo.GetAll()
}

// Audit logging
func (s *adminService) LogAction(adminID, action, resource, resourceID string, details map[string]interface{}, ipAddress, userAgent string) error {
	log := &domain.AuditLog{
		ID:         uuid.New().String(),
		AdminID:    adminID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  time.Now(),
	}

	return s.auditLogRepo.Create(log)
}

func (s *adminService) GetAuditLogs(adminID string, limit, offset int) ([]domain.AuditLog, error) {
	return s.auditLogRepo.GetByAdminID(adminID, limit, offset)
}

func (s *adminService) GetResourceAuditLogs(resource string, limit, offset int) ([]domain.AuditLog, error) {
	return s.auditLogRepo.GetByResource(resource, limit, offset)
}

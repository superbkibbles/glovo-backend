package app

import (
	"errors"
	"fmt"
	"time"

	"glovo-backend/services/driver-service/internal/domain"

	"github.com/google/uuid"
)

type driverService struct {
	driverRepo      domain.DriverRepository
	documentRepo    domain.DriverDocumentRepository
	userService     domain.UserService
	locationService domain.LocationService
	paymentService  domain.PaymentService
}

func NewDriverService(
	driverRepo domain.DriverRepository,
	documentRepo domain.DriverDocumentRepository,
	userService domain.UserService,
	locationService domain.LocationService,
	paymentService domain.PaymentService,
) domain.DriverService {
	return &driverService{
		driverRepo:      driverRepo,
		documentRepo:    documentRepo,
		userService:     userService,
		locationService: locationService,
		paymentService:  paymentService,
	}
}

func (s *driverService) RegisterDriver(userID string, req domain.RegisterDriverRequest) (*domain.Driver, error) {
	// Check if driver already exists
	if existing, _ := s.driverRepo.GetByUserID(userID); existing != nil {
		return nil, errors.New("driver profile already exists for this user")
	}

	// Validate required fields
	if req.Profile.FirstName == "" || req.Profile.LastName == "" || req.Profile.Phone == "" {
		return nil, errors.New("first name, last name, and phone are required")
	}

	driver := &domain.Driver{
		ID:       uuid.New().String(),
		UserID:   userID,
		Status:   domain.StatusOffline,
		Profile:  req.Profile,
		Vehicle:  req.Vehicle,
		BankInfo: req.BankInfo,
		Performance: domain.PerformanceStats{
			Rating:              5.0,
			TotalDeliveries:     0,
			CompletedDeliveries: 0,
			CancelledDeliveries: 0,
			TotalEarnings:       0,
		},
		Availability: domain.AvailabilityInfo{
			IsAvailable: false,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.driverRepo.Create(driver); err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	return driver, nil
}

func (s *driverService) GetDriver(driverID string) (*domain.Driver, error) {
	return s.driverRepo.GetByID(driverID)
}

func (s *driverService) GetDriverByUser(userID string) (*domain.Driver, error) {
	return s.driverRepo.GetByUserID(userID)
}

func (s *driverService) UpdateProfile(driverID string, userID string, req domain.UpdateDriverProfileRequest) (*domain.Driver, error) {
	driver, err := s.driverRepo.GetByID(driverID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if driver.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// Update profile
	if req.Profile != nil {
		driver.Profile = *req.Profile
	}

	// Update vehicle
	if req.Vehicle != nil {
		driver.Vehicle = *req.Vehicle
	}

	// Update availability
	if req.Availability != nil {
		driver.Availability = *req.Availability
	}

	// Update bank info
	if req.BankInfo != nil {
		driver.BankInfo = *req.BankInfo
	}

	driver.UpdatedAt = time.Now()

	if err := s.driverRepo.Update(driver); err != nil {
		return nil, err
	}

	return driver, nil
}

func (s *driverService) UpdateStatus(driverID string, userID string, status domain.DriverStatus) (*domain.Driver, error) {
	driver, err := s.driverRepo.GetByID(driverID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if driver.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	driver.Status = status
	driver.UpdatedAt = time.Now()

	if err := s.driverRepo.Update(driver); err != nil {
		return nil, err
	}

	return driver, nil
}

func (s *driverService) UpdateLocation(driverID string, userID string, req domain.UpdateLocationRequest) (*domain.Driver, error) {
	driver, err := s.driverRepo.GetByID(driverID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if driver.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	location := &domain.CurrentLocation{
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		UpdatedAt: time.Now(),
	}

	driver.Location = location
	driver.UpdatedAt = time.Now()

	if err := s.driverRepo.Update(driver); err != nil {
		return nil, err
	}

	// Update location service
	go s.locationService.UpdateDriverLocation(driverID, req.Latitude, req.Longitude)

	return driver, nil
}

func (s *driverService) UploadDocument(driverID string, userID string, req domain.UploadDocumentRequest) (*domain.DriverDocument, error) {
	// Verify driver exists and ownership
	driver, err := s.driverRepo.GetByID(driverID)
	if err != nil {
		return nil, err
	}

	if driver.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	document := &domain.DriverDocument{
		ID:         uuid.New().String(),
		DriverID:   driverID,
		Type:       req.Type,
		URL:        req.URL,
		Status:     domain.DocStatusPending,
		ExpiryDate: req.ExpiryDate,
		UploadedAt: time.Now(),
	}

	if err := s.documentRepo.Create(document); err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	return document, nil
}

func (s *driverService) GetDocuments(driverID string, userID string) ([]domain.DriverDocument, error) {
	// Verify driver exists and ownership
	driver, err := s.driverRepo.GetByID(driverID)
	if err != nil {
		return nil, err
	}

	if driver.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	return s.documentRepo.GetByDriverID(driverID)
}

func (s *driverService) SearchDrivers(req domain.DriverSearchRequest) ([]domain.Driver, error) {
	return s.driverRepo.Search(req)
}

func (s *driverService) GetAvailableDrivers(latitude, longitude, radius float64) ([]domain.Driver, error) {
	req := domain.DriverSearchRequest{
		Status:    domain.StatusOnline,
		Latitude:  latitude,
		Longitude: longitude,
		Radius:    radius,
		Available: &[]bool{true}[0], // pointer to true
		Limit:     50,
	}

	return s.driverRepo.Search(req)
}

func (s *driverService) GetEarningsReport(driverID string, userID string, req domain.EarningsReportRequest) ([]domain.EarningsReport, error) {
	// Verify driver exists and ownership
	driver, err := s.driverRepo.GetByID(driverID)
	if err != nil {
		return nil, err
	}

	if driver.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// Get earnings from payment service
	earningsReport, err := s.paymentService.GetDriverEarnings(driverID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	return []domain.EarningsReport{*earningsReport}, nil
}

func (s *driverService) ApproveDocument(documentID string, adminID string) error {
	document, err := s.documentRepo.GetByID(documentID)
	if err != nil {
		return err
	}

	document.Status = domain.DocStatusApproved

	return s.documentRepo.Update(document)
}

func (s *driverService) RejectDocument(documentID string, adminID string, reason string) error {
	document, err := s.documentRepo.GetByID(documentID)
	if err != nil {
		return err
	}

	document.Status = domain.DocStatusRejected

	return s.documentRepo.Update(document)
}

func (s *driverService) UpdatePerformance(driverID string, stats domain.PerformanceStats) error {
	driver, err := s.driverRepo.GetByID(driverID)
	if err != nil {
		return err
	}

	driver.Performance = stats
	driver.UpdatedAt = time.Now()

	return s.driverRepo.Update(driver)
}

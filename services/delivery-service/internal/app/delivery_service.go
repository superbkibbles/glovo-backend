package app

import (
	"errors"
	"fmt"
	"time"

	"glovo-backend/services/delivery-service/internal/domain"
	"glovo-backend/shared/auth"

	"github.com/google/uuid"
)

type deliveryService struct {
	deliveryRepo        domain.DeliveryRepository
	assignmentRepo      domain.DeliveryAssignmentRepository
	performanceRepo     domain.DriverPerformanceRepository
	orderService        domain.OrderService
	driverService       domain.DriverService
	locationService     domain.LocationService
	notificationService domain.NotificationService
	paymentService      domain.PaymentService
}

func NewDeliveryService(
	deliveryRepo domain.DeliveryRepository,
	assignmentRepo domain.DeliveryAssignmentRepository,
	performanceRepo domain.DriverPerformanceRepository,
	orderService domain.OrderService,
	driverService domain.DriverService,
	locationService domain.LocationService,
	notificationService domain.NotificationService,
	paymentService domain.PaymentService,
) domain.DeliveryService {
	return &deliveryService{
		deliveryRepo:        deliveryRepo,
		assignmentRepo:      assignmentRepo,
		performanceRepo:     performanceRepo,
		orderService:        orderService,
		driverService:       driverService,
		locationService:     locationService,
		notificationService: notificationService,
		paymentService:      paymentService,
	}
}

// Delivery management
func (s *deliveryService) CreateDelivery(req domain.CreateDeliveryRequest) (*domain.DeliveryResponse, error) {
	// Check if delivery already exists for this order
	if existing, _ := s.deliveryRepo.GetByOrderID(req.OrderID); existing != nil {
		return nil, errors.New("delivery already exists for this order")
	}

	delivery := &domain.Delivery{
		ID:              uuid.New().String(),
		OrderID:         req.OrderID,
		Status:          domain.StatusPending,
		AssignmentType:  domain.AssignmentAuto, // Default to auto assignment
		PickupAddress:   req.PickupAddress,
		DeliveryAddress: req.DeliveryAddress,
		EstimatedTime:   req.EstimatedTime,
		Distance:        req.Distance,
		DeliveryFee:     req.DeliveryFee,
		Priority:        req.Priority,
		Notes:           req.Notes,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.deliveryRepo.Create(delivery); err != nil {
		return nil, fmt.Errorf("failed to create delivery: %w", err)
	}

	// Try auto-assignment
	go s.tryAutoAssignment(delivery)

	return s.buildDeliveryResponse(delivery)
}

func (s *deliveryService) GetDelivery(deliveryID string) (*domain.DeliveryResponse, error) {
	delivery, err := s.deliveryRepo.GetByID(deliveryID)
	if err != nil {
		return nil, err
	}

	return s.buildDeliveryResponse(delivery)
}

func (s *deliveryService) GetDeliveryByOrder(orderID string) (*domain.DeliveryResponse, error) {
	delivery, err := s.deliveryRepo.GetByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	return s.buildDeliveryResponse(delivery)
}

func (s *deliveryService) UpdateDeliveryStatus(deliveryID string, req domain.UpdateDeliveryStatusRequest, userID string, role auth.UserRole) (*domain.DeliveryResponse, error) {
	delivery, err := s.deliveryRepo.GetByID(deliveryID)
	if err != nil {
		return nil, err
	}

	// Authorization check
	if role == auth.RoleDriver && (delivery.DriverID == nil || *delivery.DriverID != userID) {
		return nil, errors.New("unauthorized: driver can only update their own deliveries")
	}

	// Validate status transition
	if !s.isValidStatusTransition(delivery.Status, req.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", delivery.Status, req.Status)
	}

	// Update delivery
	delivery.Status = req.Status
	delivery.UpdatedAt = time.Now()

	// Set timestamps based on status
	now := time.Now()
	switch req.Status {
	case domain.StatusPickedUp:
		delivery.PickedUpAt = &now
	case domain.StatusDelivered:
		delivery.DeliveredAt = &now
		if delivery.PickedUpAt != nil {
			actualTime := int(now.Sub(*delivery.PickedUpAt).Minutes())
			delivery.ActualTime = &actualTime
		}
	case domain.StatusCancelled:
		delivery.CancelledAt = &now
		delivery.CancellationReason = &req.Notes
	}

	if err := s.deliveryRepo.Update(delivery); err != nil {
		return nil, err
	}

	// Update order status
	go s.updateOrderStatus(delivery.OrderID, req.Status)

	// Send notifications
	go s.sendStatusNotification(delivery, req.Status)

	// Update driver performance if delivery completed
	if req.Status == domain.StatusDelivered && delivery.DriverID != nil {
		go s.UpdateDriverPerformance(*delivery.DriverID)
	}

	return s.buildDeliveryResponse(delivery)
}

func (s *deliveryService) CancelDelivery(deliveryID string, reason string, userID string, role auth.UserRole) error {
	delivery, err := s.deliveryRepo.GetByID(deliveryID)
	if err != nil {
		return err
	}

	// Authorization check
	if role == auth.RoleDriver && (delivery.DriverID == nil || *delivery.DriverID != userID) {
		return errors.New("unauthorized")
	}

	if delivery.Status == domain.StatusDelivered {
		return errors.New("cannot cancel delivered order")
	}

	delivery.Status = domain.StatusCancelled
	now := time.Now()
	delivery.CancelledAt = &now
	delivery.CancellationReason = &reason
	delivery.UpdatedAt = now

	if err := s.deliveryRepo.Update(delivery); err != nil {
		return err
	}

	// Update order status
	go s.updateOrderStatus(delivery.OrderID, domain.StatusCancelled)

	// Send notifications
	go s.sendStatusNotification(delivery, domain.StatusCancelled)

	return nil
}

// Driver assignment
func (s *deliveryService) AutoAssignDriver(req domain.AutoAssignmentRequest) (*domain.DeliveryResponse, error) {
	delivery, err := s.deliveryRepo.GetByID(req.DeliveryID)
	if err != nil {
		return nil, err
	}

	if delivery.Status != domain.StatusPending {
		return nil, errors.New("delivery is not in pending status")
	}

	// Get available drivers
	drivers, err := s.GetAvailableDrivers(req.Latitude, req.Longitude, req.Radius)
	if err != nil {
		return nil, err
	}

	if len(drivers) == 0 {
		return nil, errors.New("no available drivers found")
	}

	// Select best driver (closest with highest rating)
	bestDriver := s.selectBestDriver(drivers)

	// Create assignment
	assignment := &domain.DeliveryAssignment{
		ID:         uuid.New().String(),
		DeliveryID: req.DeliveryID,
		DriverID:   bestDriver.DriverID,
		Status:     domain.AssignmentPending,
		ExpiresAt:  time.Now().Add(5 * time.Minute), // 5 minute expiry
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.assignmentRepo.Create(assignment); err != nil {
		return nil, err
	}

	// Update delivery status
	delivery.Status = domain.StatusAssigned
	delivery.DriverID = &bestDriver.DriverID
	now := time.Now()
	delivery.AssignedAt = &now
	delivery.UpdatedAt = now

	if err := s.deliveryRepo.Update(delivery); err != nil {
		return nil, err
	}

	// Send notification to driver
	go s.notificationService.SendDeliveryAssignment(bestDriver.DriverID, delivery)

	return s.buildDeliveryResponse(delivery)
}

func (s *deliveryService) ManualAssignDriver(req domain.AssignDriverRequest, adminID string) (*domain.DeliveryResponse, error) {
	delivery, err := s.deliveryRepo.GetByID(req.DeliveryID)
	if err != nil {
		return nil, err
	}

	// Check if driver is available
	available, err := s.driverService.IsDriverAvailable(req.DriverID)
	if err != nil {
		return nil, err
	}

	if !available {
		return nil, errors.New("driver is not available")
	}

	// Create assignment
	assignment := &domain.DeliveryAssignment{
		ID:         uuid.New().String(),
		DeliveryID: req.DeliveryID,
		DriverID:   req.DriverID,
		Status:     domain.AssignmentPending,
		ExpiresAt:  time.Now().Add(10 * time.Minute), // 10 minute expiry for manual assignment
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.assignmentRepo.Create(assignment); err != nil {
		return nil, err
	}

	// Update delivery
	delivery.Status = domain.StatusAssigned
	delivery.DriverID = &req.DriverID
	delivery.AssignmentType = req.Type
	now := time.Now()
	delivery.AssignedAt = &now
	delivery.UpdatedAt = now

	if err := s.deliveryRepo.Update(delivery); err != nil {
		return nil, err
	}

	// Send notification to driver
	go s.notificationService.SendDeliveryAssignment(req.DriverID, delivery)

	return s.buildDeliveryResponse(delivery)
}

func (s *deliveryService) GetAvailableDrivers(latitude, longitude, radius float64) ([]domain.DriverAvailability, error) {
	return s.driverService.GetAvailableDrivers(latitude, longitude, radius)
}

// Driver responses
func (s *deliveryService) RespondToAssignment(driverID string, req domain.DriverResponseRequest) error {
	// Get the assignment
	assignments, err := s.assignmentRepo.GetByDeliveryID(req.DeliveryID)
	if err != nil {
		return err
	}

	var assignment *domain.DeliveryAssignment
	for _, a := range assignments {
		if a.DriverID == driverID && a.Status == domain.AssignmentPending {
			assignment = &a
			break
		}
	}

	if assignment == nil {
		return errors.New("no pending assignment found for this driver")
	}

	// Check if assignment has expired
	if time.Now().After(assignment.ExpiresAt) {
		assignment.Status = domain.AssignmentExpired
		s.assignmentRepo.Update(assignment)
		return errors.New("assignment has expired")
	}

	// Update assignment
	assignment.Response = &domain.AssignmentResponse{
		ResponseTime: time.Now(),
		Reason:       req.Reason,
	}
	assignment.UpdatedAt = time.Now()

	delivery, err := s.deliveryRepo.GetByID(req.DeliveryID)
	if err != nil {
		return err
	}

	if req.Accept {
		assignment.Status = domain.AssignmentAccepted
		delivery.Status = domain.StatusAccepted

		// Update driver status
		go s.driverService.UpdateDriverStatus(driverID, "busy")

		// Create delivery route
		pickupLocation := domain.Location{
			Latitude:  delivery.PickupAddress.Latitude,
			Longitude: delivery.PickupAddress.Longitude,
			Timestamp: time.Now(),
		}
		dropoffLocation := domain.Location{
			Latitude:  delivery.DeliveryAddress.Latitude,
			Longitude: delivery.DeliveryAddress.Longitude,
			Timestamp: time.Now(),
		}
		go s.locationService.CreateDeliveryRoute(delivery.ID, driverID, pickupLocation, dropoffLocation)
	} else {
		assignment.Status = domain.AssignmentRejected
		delivery.Status = domain.StatusPending
		delivery.DriverID = nil
		delivery.AssignedAt = nil

		// Try to reassign automatically
		go s.tryAutoAssignment(delivery)
	}

	if err := s.assignmentRepo.Update(assignment); err != nil {
		return err
	}

	delivery.UpdatedAt = time.Now()
	if err := s.deliveryRepo.Update(delivery); err != nil {
		return err
	}

	// Send notifications
	go s.sendStatusNotification(delivery, delivery.Status)

	return nil
}

func (s *deliveryService) GetDriverAssignments(driverID string, limit, offset int) ([]domain.DeliveryAssignment, error) {
	return s.assignmentRepo.GetByDriverID(driverID, limit, offset)
}

func (s *deliveryService) GetPendingAssignments(driverID string) ([]domain.DeliveryAssignment, error) {
	return s.assignmentRepo.GetPendingForDriver(driverID)
}

// Driver operations
func (s *deliveryService) PickupOrder(deliveryID string, driverID string) (*domain.DeliveryResponse, error) {
	delivery, err := s.deliveryRepo.GetByID(deliveryID)
	if err != nil {
		return nil, err
	}

	if delivery.DriverID == nil || *delivery.DriverID != driverID {
		return nil, errors.New("unauthorized")
	}

	if delivery.Status != domain.StatusAccepted {
		return nil, errors.New("delivery must be accepted before pickup")
	}

	delivery.Status = domain.StatusPickedUp
	now := time.Now()
	delivery.PickedUpAt = &now
	delivery.UpdatedAt = now

	if err := s.deliveryRepo.Update(delivery); err != nil {
		return nil, err
	}

	// Update order status
	go s.updateOrderStatus(delivery.OrderID, domain.StatusPickedUp)

	// Send notifications
	go s.sendStatusNotification(delivery, domain.StatusPickedUp)

	return s.buildDeliveryResponse(delivery)
}

func (s *deliveryService) CompleteDelivery(deliveryID string, driverID string) (*domain.DeliveryResponse, error) {
	delivery, err := s.deliveryRepo.GetByID(deliveryID)
	if err != nil {
		return nil, err
	}

	if delivery.DriverID == nil || *delivery.DriverID != driverID {
		return nil, errors.New("unauthorized")
	}

	if delivery.Status != domain.StatusPickedUp && delivery.Status != domain.StatusInTransit {
		return nil, errors.New("delivery must be picked up before completion")
	}

	delivery.Status = domain.StatusDelivered
	now := time.Now()
	delivery.DeliveredAt = &now

	if delivery.PickedUpAt != nil {
		actualTime := int(now.Sub(*delivery.PickedUpAt).Minutes())
		delivery.ActualTime = &actualTime
	}

	delivery.UpdatedAt = now

	if err := s.deliveryRepo.Update(delivery); err != nil {
		return nil, err
	}

	// Update driver status back to online
	go s.driverService.UpdateDriverStatus(driverID, "online")

	// Process payment
	go s.paymentService.ProcessDeliveryPayment(deliveryID)

	// Update order status
	go s.updateOrderStatus(delivery.OrderID, domain.StatusDelivered)

	// Send notifications
	go s.sendStatusNotification(delivery, domain.StatusDelivered)

	// Update driver performance
	go s.UpdateDriverPerformance(driverID)

	return s.buildDeliveryResponse(delivery)
}

func (s *deliveryService) ReportIssue(deliveryID string, driverID string, issue string) error {
	delivery, err := s.deliveryRepo.GetByID(deliveryID)
	if err != nil {
		return err
	}

	if delivery.DriverID == nil || *delivery.DriverID != driverID {
		return errors.New("unauthorized")
	}

	// For now, just send notification to support
	go s.notificationService.SendDriverNotification("admin", fmt.Sprintf("Issue reported for delivery %s: %s", deliveryID, issue))

	return nil
}

// Analytics and metrics
func (s *deliveryService) GetDeliveryMetrics() (*domain.DeliveryMetrics, error) {
	// This is a simplified implementation
	// In production, you'd want to use aggregation queries

	pendingDeliveries, _ := s.deliveryRepo.GetPendingDeliveries()
	activeDeliveries, _ := s.deliveryRepo.GetActiveDeliveries()

	return &domain.DeliveryMetrics{
		PendingDeliveries: len(pendingDeliveries),
		ActiveDeliveries:  len(activeDeliveries),
		// Other metrics would be calculated from database aggregations
	}, nil
}

func (s *deliveryService) GetDriverPerformance(driverID string) (*domain.DriverPerformance, error) {
	return s.performanceRepo.GetByDriverID(driverID)
}

func (s *deliveryService) UpdateDriverPerformance(driverID string) error {
	// Get all deliveries for this driver
	deliveries, err := s.deliveryRepo.GetByDriverID(driverID, 1000, 0)
	if err != nil {
		return err
	}

	// Calculate performance metrics
	performance := &domain.DriverPerformance{
		DriverID:    driverID,
		LastUpdated: time.Now(),
	}

	var totalDeliveryTime float64
	var onTimeCount int
	var completedCount int

	for _, delivery := range deliveries {
		performance.TotalDeliveries++

		if delivery.Status == domain.StatusDelivered {
			completedCount++

			if delivery.ActualTime != nil && delivery.EstimatedTime > 0 {
				totalDeliveryTime += float64(*delivery.ActualTime)

				// Consider on-time if within 10% of estimated time
				if *delivery.ActualTime <= int(float64(delivery.EstimatedTime)*1.1) {
					onTimeCount++
				}
			}
		} else if delivery.Status == domain.StatusCancelled {
			performance.CancelledDeliveries++
		}
	}

	performance.CompletedDeliveries = completedCount

	if completedCount > 0 {
		performance.AverageDeliveryTime = totalDeliveryTime / float64(completedCount)
		performance.OnTimeDeliveryRate = float64(onTimeCount) / float64(completedCount) * 100
	}

	// Get existing performance to preserve ID
	existing, err := s.performanceRepo.GetByDriverID(driverID)
	if err == nil {
		performance.ID = existing.ID
		return s.performanceRepo.Update(performance)
	} else {
		performance.ID = uuid.New().String()
		return s.performanceRepo.Create(performance)
	}
}

func (s *deliveryService) GetDriverRankings() ([]domain.DriverPerformance, error) {
	return s.performanceRepo.GetDriverRankings()
}

// Admin operations
func (s *deliveryService) SearchDeliveries(req domain.DeliverySearchRequest) ([]domain.Delivery, error) {
	return s.deliveryRepo.Search(req)
}

func (s *deliveryService) ReassignDelivery(deliveryID string, newDriverID string, adminID string) error {
	delivery, err := s.deliveryRepo.GetByID(deliveryID)
	if err != nil {
		return err
	}

	// Check if new driver is available
	available, err := s.driverService.IsDriverAvailable(newDriverID)
	if err != nil {
		return err
	}

	if !available {
		return errors.New("new driver is not available")
	}

	// Update current driver status if assigned
	if delivery.DriverID != nil {
		go s.driverService.UpdateDriverStatus(*delivery.DriverID, "online")
	}

	// Update delivery
	delivery.DriverID = &newDriverID
	delivery.Status = domain.StatusAssigned
	now := time.Now()
	delivery.AssignedAt = &now
	delivery.UpdatedAt = now

	if err := s.deliveryRepo.Update(delivery); err != nil {
		return err
	}

	// Create new assignment
	assignment := &domain.DeliveryAssignment{
		ID:         uuid.New().String(),
		DeliveryID: deliveryID,
		DriverID:   newDriverID,
		Status:     domain.AssignmentPending,
		ExpiresAt:  time.Now().Add(10 * time.Minute),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.assignmentRepo.Create(assignment); err != nil {
		return err
	}

	// Send notification to new driver
	go s.notificationService.SendDeliveryAssignment(newDriverID, delivery)

	return nil
}

func (s *deliveryService) GetSystemStats() (*domain.DeliveryMetrics, error) {
	return s.GetDeliveryMetrics()
}

// Helper functions
func (s *deliveryService) buildDeliveryResponse(delivery *domain.Delivery) (*domain.DeliveryResponse, error) {
	response := &domain.DeliveryResponse{
		Delivery: *delivery,
	}

	// Get driver info if assigned
	if delivery.DriverID != nil {
		if driverInfo, err := s.driverService.GetDriver(*delivery.DriverID); err == nil {
			response.DriverInfo = driverInfo
		}
	}

	// Get order info
	if orderInfo, err := s.orderService.GetOrder(delivery.OrderID); err == nil {
		response.OrderInfo = orderInfo
	}

	// Get tracking info
	if tracking, err := s.locationService.GetDeliveryTracking(delivery.ID); err == nil {
		response.TrackingInfo = tracking
	}

	return response, nil
}

func (s *deliveryService) isValidStatusTransition(from, to domain.DeliveryStatus) bool {
	validTransitions := map[domain.DeliveryStatus][]domain.DeliveryStatus{
		domain.StatusPending:   {domain.StatusAssigned, domain.StatusCancelled},
		domain.StatusAssigned:  {domain.StatusAccepted, domain.StatusRejected, domain.StatusCancelled},
		domain.StatusAccepted:  {domain.StatusPickedUp, domain.StatusCancelled},
		domain.StatusPickedUp:  {domain.StatusInTransit, domain.StatusDelivered, domain.StatusFailed},
		domain.StatusInTransit: {domain.StatusDelivered, domain.StatusFailed},
	}

	allowed, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == to {
			return true
		}
	}

	return false
}

func (s *deliveryService) selectBestDriver(drivers []domain.DriverAvailability) domain.DriverAvailability {
	if len(drivers) == 0 {
		return domain.DriverAvailability{}
	}

	// Score drivers based on distance and rating
	bestScore := -1.0
	bestIndex := 0

	for i, driver := range drivers {
		// Normalize distance (closer is better)
		distanceScore := 1.0 / (1.0 + driver.Distance)

		// Rating score (0-5 scale)
		ratingScore := driver.Rating / 5.0

		// Combined score (70% distance, 30% rating)
		score := 0.7*distanceScore + 0.3*ratingScore

		if score > bestScore {
			bestScore = score
			bestIndex = i
		}
	}

	return drivers[bestIndex]
}

func (s *deliveryService) tryAutoAssignment(delivery *domain.Delivery) {
	req := domain.AutoAssignmentRequest{
		DeliveryID: delivery.ID,
		Latitude:   delivery.PickupAddress.Latitude,
		Longitude:  delivery.PickupAddress.Longitude,
		Radius:     10.0, // 10km radius
	}

	_, err := s.AutoAssignDriver(req)
	if err != nil {
		// Log error or handle failed auto-assignment
		// Could try with larger radius or escalate to manual assignment
	}
}

func (s *deliveryService) updateOrderStatus(orderID string, status domain.DeliveryStatus) {
	// Map delivery status to order status
	orderStatus := "in_progress"
	switch status {
	case domain.StatusPickedUp:
		orderStatus = "picked_up"
	case domain.StatusInTransit:
		orderStatus = "in_transit"
	case domain.StatusDelivered:
		orderStatus = "delivered"
	case domain.StatusCancelled:
		orderStatus = "cancelled"
	}

	s.orderService.UpdateOrderStatus(orderID, orderStatus)
}

func (s *deliveryService) sendStatusNotification(delivery *domain.Delivery, status domain.DeliveryStatus) {
	s.notificationService.SendDeliveryUpdate(delivery.OrderID, status)
}

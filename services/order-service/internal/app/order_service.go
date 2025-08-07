package app

import (
	"errors"
	"fmt"
	"time"

	"glovo-backend/services/order-service/internal/domain"
	"glovo-backend/shared/auth"

	"github.com/google/uuid"
)

type orderService struct {
	orderRepo           domain.OrderRepository
	catalogService      domain.CatalogService
	paymentService      domain.PaymentService
	notificationService domain.NotificationService
}

func NewOrderService(
	orderRepo domain.OrderRepository,
	catalogService domain.CatalogService,
	paymentService domain.PaymentService,
	notificationService domain.NotificationService,
) domain.OrderService {
	return &orderService{
		orderRepo:           orderRepo,
		catalogService:      catalogService,
		paymentService:      paymentService,
		notificationService: notificationService,
	}
}

func (s *orderService) CreateOrder(customerID string, req domain.CreateOrderRequest) (*domain.OrderResponse, error) {
	// Validate order with catalog service
	validation, err := s.catalogService.ValidateOrder(req.MerchantID, req.Items)
	if err != nil {
		return nil, fmt.Errorf("failed to validate order: %w", err)
	}

	if !validation.Valid {
		return nil, fmt.Errorf("order validation failed: %v", validation.Errors)
	}

	// Create order entity
	order := &domain.Order{
		ID:           uuid.New().String(),
		CustomerID:   customerID,
		MerchantID:   req.MerchantID,
		Status:       domain.StatusPending,
		DeliveryInfo: req.DeliveryInfo,
		PaymentInfo:  req.PaymentInfo,
		TotalAmount:  validation.TotalAmount,
		DeliveryFee:  calculateDeliveryFee(req.DeliveryInfo),
		ServiceFee:   calculateServiceFee(validation.TotalAmount),
		TaxAmount:    calculateTax(validation.TotalAmount),
		PlacedAt:     time.Now(),
		ScheduledFor: req.ScheduledFor,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Calculate final amount
	order.FinalAmount = order.TotalAmount + order.DeliveryFee + order.ServiceFee + order.TaxAmount

	// Create order items
	for _, validatedItem := range validation.Items {
		item := domain.OrderItem{
			ID:        uuid.New().String(),
			OrderID:   order.ID,
			ProductID: validatedItem.ProductID,
			Name:      validatedItem.Name,
			Price:     validatedItem.Price,
			Quantity:  validatedItem.Quantity,
		}

		// Find corresponding notes from request
		for _, reqItem := range req.Items {
			if reqItem.ProductID == validatedItem.ProductID {
				item.Notes = reqItem.Notes
				break
			}
		}

		order.Items = append(order.Items, item)
	}

	// Process payment
	paymentResult, err := s.paymentService.ProcessPayment(order.ID, order.FinalAmount, req.PaymentInfo)
	if err != nil {
		return nil, fmt.Errorf("payment processing failed: %w", err)
	}

	if !paymentResult.Success {
		return nil, fmt.Errorf("payment failed: %s", paymentResult.Error)
	}

	// Update payment info with result
	order.PaymentInfo.Status = "completed"
	order.PaymentInfo.Reference = paymentResult.Reference

	// Save order to database
	if err := s.orderRepo.Create(order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Send notification
	message := fmt.Sprintf("Order #%s has been placed successfully", order.ID[:8])
	s.notificationService.SendOrderNotification(order.ID, customerID, message)

	// Emit OrderCreated event (placeholder)
	s.emitOrderCreatedEvent(order)

	return &domain.OrderResponse{
		Order: *order,
	}, nil
}

func (s *orderService) GetOrder(orderID string, userID string, role auth.UserRole) (*domain.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if !s.canAccessOrder(order, userID, role) {
		return nil, errors.New("unauthorized access to order")
	}

	response := &domain.OrderResponse{
		Order: *order,
	}

	// Add additional info based on role
	if role == auth.RoleCustomer || role == auth.RoleAdmin {
		// Add merchant info (would call merchant service)
		response.MerchantInfo = &domain.MerchantInfo{
			ID:   order.MerchantID,
			Name: "Sample Restaurant", // Mock data
		}

		// Add driver info if assigned
		if order.DriverID != nil {
			response.DriverInfo = &domain.DriverInfo{
				ID:   *order.DriverID,
				Name: "Sample Driver", // Mock data
			}
		}

		// Add tracking info
		response.TrackingInfo = s.buildTrackingInfo(order)
	}

	return response, nil
}

func (s *orderService) GetOrderHistory(userID string, role auth.UserRole, limit, offset int) ([]domain.Order, error) {
	switch role {
	case auth.RoleCustomer:
		return s.orderRepo.GetByCustomerID(userID, limit, offset)
	case auth.RoleMerchant:
		return s.orderRepo.GetByMerchantID(userID, limit, offset)
	case auth.RoleDriver:
		return s.orderRepo.GetByDriverID(userID, limit, offset)
	case auth.RoleAdmin:
		return s.orderRepo.List(limit, offset)
	default:
		return nil, errors.New("invalid role")
	}
}

func (s *orderService) UpdateOrderStatus(orderID string, req domain.UpdateOrderStatusRequest, userID string, role auth.UserRole) (*domain.OrderResponse, error) {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, err
	}

	// Check authorization for status updates
	if !s.canUpdateOrderStatus(order, userID, role, req.Status) {
		return nil, errors.New("unauthorized to update order status")
	}

	// Validate status transition
	if !s.isValidStatusTransition(order.Status, req.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", order.Status, req.Status)
	}

	// Update order
	order.Status = req.Status
	order.UpdatedAt = time.Now()

	if req.EstimatedTime != nil {
		order.EstimatedTime = req.EstimatedTime
	}

	if req.Status == domain.StatusDelivered {
		now := time.Now()
		order.CompletedAt = &now
	}

	if req.Status == domain.StatusCancelled {
		now := time.Now()
		order.CancelledAt = &now
		if req.CancellationReason != nil {
			order.CancellationReason = req.CancellationReason
		}
	}

	if err := s.orderRepo.Update(order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	// Send notification
	message := fmt.Sprintf("Order #%s status updated to %s", order.ID[:8], req.Status)
	s.notificationService.SendOrderNotification(order.ID, order.CustomerID, message)

	return s.GetOrder(orderID, userID, role)
}

func (s *orderService) CancelOrder(orderID string, userID string, role auth.UserRole, reason string) (*domain.OrderResponse, error) {
	req := domain.UpdateOrderStatusRequest{
		Status:             domain.StatusCancelled,
		CancellationReason: &reason,
	}
	return s.UpdateOrderStatus(orderID, req, userID, role)
}

func (s *orderService) GetOrdersForMerchant(merchantID string, limit, offset int) ([]domain.Order, error) {
	return s.orderRepo.GetByMerchantID(merchantID, limit, offset)
}

func (s *orderService) GetOrdersForDriver(driverID string, limit, offset int) ([]domain.Order, error) {
	return s.orderRepo.GetByDriverID(driverID, limit, offset)
}

func (s *orderService) GetActiveOrders() ([]domain.Order, error) {
	// Get orders that are not completed or cancelled
	var activeOrders []domain.Order

	statuses := []domain.OrderStatus{
		domain.StatusPending,
		domain.StatusConfirmed,
		domain.StatusPreparing,
		domain.StatusReady,
		domain.StatusAssigned,
		domain.StatusPickedUp,
		domain.StatusInTransit,
	}

	for _, status := range statuses {
		orders, err := s.orderRepo.GetByStatus(status, 100, 0)
		if err != nil {
			return nil, err
		}
		activeOrders = append(activeOrders, orders...)
	}

	return activeOrders, nil
}

// Helper functions
func (s *orderService) canAccessOrder(order *domain.Order, userID string, role auth.UserRole) bool {
	switch role {
	case auth.RoleCustomer:
		return order.CustomerID == userID
	case auth.RoleMerchant:
		return order.MerchantID == userID
	case auth.RoleDriver:
		return order.DriverID != nil && *order.DriverID == userID
	case auth.RoleAdmin:
		return true
	default:
		return false
	}
}

func (s *orderService) canUpdateOrderStatus(order *domain.Order, userID string, role auth.UserRole, newStatus domain.OrderStatus) bool {
	switch role {
	case auth.RoleMerchant:
		// Merchants can update to confirmed, preparing, ready
		return order.MerchantID == userID &&
			(newStatus == domain.StatusConfirmed ||
				newStatus == domain.StatusPreparing ||
				newStatus == domain.StatusReady ||
				newStatus == domain.StatusCancelled)
	case auth.RoleDriver:
		// Drivers can update to picked_up, in_transit, delivered
		return order.DriverID != nil && *order.DriverID == userID &&
			(newStatus == domain.StatusPickedUp ||
				newStatus == domain.StatusInTransit ||
				newStatus == domain.StatusDelivered)
	case auth.RoleAdmin:
		return true
	default:
		return false
	}
}

func (s *orderService) isValidStatusTransition(from, to domain.OrderStatus) bool {
	validTransitions := map[domain.OrderStatus][]domain.OrderStatus{
		domain.StatusPending:   {domain.StatusConfirmed, domain.StatusCancelled},
		domain.StatusConfirmed: {domain.StatusPreparing, domain.StatusCancelled},
		domain.StatusPreparing: {domain.StatusReady, domain.StatusCancelled},
		domain.StatusReady:     {domain.StatusAssigned, domain.StatusCancelled},
		domain.StatusAssigned:  {domain.StatusPickedUp, domain.StatusCancelled},
		domain.StatusPickedUp:  {domain.StatusInTransit},
		domain.StatusInTransit: {domain.StatusDelivered},
	}

	allowedNext, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedNext {
		if allowed == to {
			return true
		}
	}

	return false
}

func (s *orderService) buildTrackingInfo(order *domain.Order) *domain.OrderTrackingInfo {
	// Mock tracking info - in real implementation, this would call location service
	steps := []domain.TrackingStep{
		{
			Status:      domain.StatusPending,
			Description: "Order placed",
			Timestamp:   order.PlacedAt,
		},
	}

	if order.Status != domain.StatusPending {
		steps = append(steps, domain.TrackingStep{
			Status:      order.Status,
			Description: getStatusDescription(order.Status),
			Timestamp:   order.UpdatedAt,
		})
	}

	return &domain.OrderTrackingInfo{
		Steps: steps,
	}
}

func (s *orderService) emitOrderCreatedEvent(order *domain.Order) {
	// Placeholder for event emission
	// In real implementation, this would publish to message queue
	fmt.Printf("Event: OrderCreated - OrderID: %s, CustomerID: %s, MerchantID: %s\n",
		order.ID, order.CustomerID, order.MerchantID)
}

func calculateDeliveryFee(deliveryInfo domain.DeliveryInfo) float64 {
	// Mock calculation - in real implementation, this would consider distance, time, etc.
	return 2.99
}

func calculateServiceFee(totalAmount float64) float64 {
	// Mock calculation - typically a percentage of total
	return totalAmount * 0.05 // 5% service fee
}

func calculateTax(totalAmount float64) float64 {
	// Mock calculation - typically a percentage
	return totalAmount * 0.08 // 8% tax
}

func getStatusDescription(status domain.OrderStatus) string {
	descriptions := map[domain.OrderStatus]string{
		domain.StatusPending:   "Order placed",
		domain.StatusConfirmed: "Order confirmed by restaurant",
		domain.StatusPreparing: "Restaurant is preparing your order",
		domain.StatusReady:     "Order is ready for pickup",
		domain.StatusAssigned:  "Driver assigned",
		domain.StatusPickedUp:  "Order picked up by driver",
		domain.StatusInTransit: "Order is on the way",
		domain.StatusDelivered: "Order delivered",
		domain.StatusCancelled: "Order cancelled",
	}

	if desc, exists := descriptions[status]; exists {
		return desc
	}
	return string(status)
}

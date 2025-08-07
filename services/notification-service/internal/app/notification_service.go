package app

import (
	"errors"
	"fmt"
	"time"

	"glovo-backend/services/notification-service/internal/domain"

	"github.com/google/uuid"
)

type notificationService struct {
	notificationRepo domain.NotificationRepository
	templateRepo     domain.TemplateRepository
	preferenceRepo   domain.PreferenceRepository
	deviceRepo       domain.DeviceRepository
	pushService      domain.PushNotificationService
	smsService       domain.SMSService
	emailService     domain.EmailService
}

func NewNotificationService(
	notificationRepo domain.NotificationRepository,
	templateRepo domain.TemplateRepository,
	preferenceRepo domain.PreferenceRepository,
	deviceRepo domain.DeviceRepository,
	pushService domain.PushNotificationService,
	smsService domain.SMSService,
	emailService domain.EmailService,
) domain.NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
		preferenceRepo:   preferenceRepo,
		deviceRepo:       deviceRepo,
		pushService:      pushService,
		smsService:       smsService,
		emailService:     emailService,
	}
}

// Sending notifications
func (s *notificationService) SendNotification(req domain.SendNotificationRequest) (*domain.Notification, error) {
	// Check user preferences for this type/channel
	preference, err := s.preferenceRepo.GetByUserAndType(req.UserID, req.Type)
	if err != nil {
		// Create default preference if none exist
		preference = &domain.UserPreference{
			ID:        uuid.New().String(),
			UserID:    req.UserID,
			Type:      req.Type,
			Channel:   req.Channel,
			Enabled:   true, // Default to enabled
			UpdatedAt: time.Now(),
		}
		s.preferenceRepo.Create(preference)
	}

	// Create notification record
	notification := &domain.Notification{
		ID:           uuid.New().String(),
		UserID:       req.UserID,
		Type:         req.Type,
		Channel:      req.Channel,
		Title:        req.Title,
		Message:      req.Message,
		Data:         req.Data,
		Status:       domain.StatusPending,
		Priority:     req.Priority,
		ScheduledFor: req.ScheduledFor,
		ExpiresAt:    req.ExpiresAt,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.notificationRepo.Create(notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Send notification based on channel and preferences
	if preference.Enabled {
		go s.processNotification(notification)
	}

	return notification, nil
}

func (s *notificationService) SendBulkNotification(req domain.BulkNotificationRequest) ([]domain.Notification, error) {
	var notifications []domain.Notification

	for _, userID := range req.UserIDs {
		notifReq := domain.SendNotificationRequest{
			UserID:       userID,
			Type:         req.Type,
			Channel:      req.Channel,
			Title:        req.Title,
			Message:      req.Message,
			Data:         req.Data,
			Priority:     req.Priority,
			ScheduledFor: req.ScheduledFor,
		}

		notification, err := s.SendNotification(notifReq)
		if err != nil {
			// Log error but continue with bulk operation
			fmt.Printf("Failed to send notification to user %s: %v\n", userID, err)
			continue
		}
		notifications = append(notifications, *notification)
	}

	return notifications, nil
}

func (s *notificationService) SendTemplateNotification(req domain.SendTemplateNotificationRequest) (*domain.Notification, error) {
	// Get template
	template, err := s.templateRepo.GetByID(req.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	// Replace variables in template
	title := s.replaceVariables(template.Title, req.Variables)
	message := s.replaceVariables(template.Message, req.Variables)

	// Send notification
	notifReq := domain.SendNotificationRequest{
		UserID:       req.UserID,
		Type:         template.Type,
		Channel:      template.Channel,
		Title:        title,
		Message:      message,
		ScheduledFor: req.ScheduledFor,
	}

	return s.SendNotification(notifReq)
}

// OTP notifications (for User Service integration)
func (s *notificationService) SendOTPNotification(phoneNumber, otpCode string) error {
	message := fmt.Sprintf("Your OTP code is: %s. Valid for 5 minutes.", otpCode)
	return s.smsService.SendSMS(phoneNumber, message)
}

// Order notifications (for Order Service integration)
func (s *notificationService) SendOrderNotification(orderID, userID, message string) error {
	req := domain.SendNotificationRequest{
		UserID:   userID,
		Type:     domain.TypeOrderUpdate,
		Channel:  domain.ChannelPush,
		Title:    "Order Update",
		Message:  message,
		Data:     map[string]string{"order_id": orderID},
		Priority: domain.PriorityNormal,
	}

	_, err := s.SendNotification(req)
	return err
}

// Managing notifications
func (s *notificationService) GetNotifications(userID string, limit, offset int) (*domain.NotificationListResponse, error) {
	notifications, err := s.notificationRepo.GetByUserID(userID, limit, offset)
	if err != nil {
		return nil, err
	}

	unreadCount, err := s.notificationRepo.GetUnreadCount(userID)
	if err != nil {
		return nil, err
	}

	return &domain.NotificationListResponse{
		Notifications: notifications,
		UnreadCount:   unreadCount,
		TotalCount:    len(notifications),
	}, nil
}

func (s *notificationService) GetNotification(notificationID string) (*domain.Notification, error) {
	return s.notificationRepo.GetByID(notificationID)
}

func (s *notificationService) MarkNotificationAsRead(notificationID, userID string) error {
	notification, err := s.notificationRepo.GetByID(notificationID)
	if err != nil {
		return err
	}

	if notification.UserID != userID {
		return errors.New("unauthorized")
	}

	notification.Status = domain.StatusRead
	notification.ReadAt = &time.Time{}
	*notification.ReadAt = time.Now()
	notification.UpdatedAt = time.Now()

	return s.notificationRepo.Update(notification)
}

func (s *notificationService) MarkAllNotificationsAsRead(userID string) error {
	return s.notificationRepo.MarkAllAsRead(userID)
}

func (s *notificationService) GetUnreadCount(userID string) (int, error) {
	return s.notificationRepo.GetUnreadCount(userID)
}

// Templates
func (s *notificationService) CreateTemplate(template *domain.NotificationTemplate) (*domain.NotificationTemplate, error) {
	template.ID = uuid.New().String()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	if err := s.templateRepo.Create(template); err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return template, nil
}

func (s *notificationService) GetTemplate(templateID string) (*domain.NotificationTemplate, error) {
	return s.templateRepo.GetByID(templateID)
}

func (s *notificationService) UpdateTemplate(templateID string, updates map[string]interface{}) (*domain.NotificationTemplate, error) {
	template, err := s.templateRepo.GetByID(templateID)
	if err != nil {
		return nil, err
	}

	// Update fields based on updates map
	if name, ok := updates["name"].(string); ok {
		template.Name = name
	}
	if title, ok := updates["title"].(string); ok {
		template.Title = title
	}
	if message, ok := updates["message"].(string); ok {
		template.Message = message
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		template.IsActive = isActive
	}

	template.UpdatedAt = time.Now()

	if err := s.templateRepo.Update(template); err != nil {
		return nil, err
	}

	return template, nil
}

func (s *notificationService) DeleteTemplate(templateID string) error {
	return s.templateRepo.Delete(templateID)
}

func (s *notificationService) ListTemplates(limit, offset int) ([]domain.NotificationTemplate, error) {
	return s.templateRepo.List(limit, offset)
}

// User preferences
func (s *notificationService) GetUserPreferences(userID string) ([]domain.UserPreference, error) {
	return s.preferenceRepo.GetByUserID(userID)
}

func (s *notificationService) UpdateUserPreference(userID string, req domain.UpdatePreferenceRequest) (*domain.UserPreference, error) {
	// Check if preference exists
	preference, err := s.preferenceRepo.GetByUserAndType(userID, req.Type)
	if err != nil {
		// Create new preference
		preference = &domain.UserPreference{
			ID:        uuid.New().String(),
			UserID:    userID,
			Type:      req.Type,
			Channel:   req.Channel,
			Enabled:   req.Enabled,
			UpdatedAt: time.Now(),
		}

		if err := s.preferenceRepo.Create(preference); err != nil {
			return nil, err
		}
	} else {
		// Update existing preference
		preference.Enabled = req.Enabled
		preference.UpdatedAt = time.Now()

		if err := s.preferenceRepo.Update(preference); err != nil {
			return nil, err
		}
	}

	return preference, nil
}

// Device management
func (s *notificationService) RegisterDevice(userID string, req domain.RegisterDeviceRequest) (*domain.NotificationDevice, error) {
	// Check if device already exists
	existing, _ := s.deviceRepo.GetByToken(req.DeviceToken)
	if existing != nil {
		// Update existing device
		existing.UserID = userID
		existing.Platform = req.Platform
		existing.AppVersion = req.AppVersion
		existing.IsActive = true
		existing.LastActiveAt = time.Now()
		existing.UpdatedAt = time.Now()

		if err := s.deviceRepo.Update(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	// Create new device
	device := &domain.NotificationDevice{
		ID:           uuid.New().String(),
		UserID:       userID,
		DeviceToken:  req.DeviceToken,
		Platform:     req.Platform,
		AppVersion:   req.AppVersion,
		IsActive:     true,
		LastActiveAt: time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.deviceRepo.Create(device); err != nil {
		return nil, fmt.Errorf("failed to register device: %w", err)
	}

	return device, nil
}

func (s *notificationService) GetUserDevices(userID string) ([]domain.NotificationDevice, error) {
	return s.deviceRepo.GetByUserID(userID)
}

func (s *notificationService) DeactivateDevice(userID, deviceToken string) error {
	device, err := s.deviceRepo.GetByToken(deviceToken)
	if err != nil {
		return err
	}

	if device.UserID != userID {
		return errors.New("unauthorized")
	}

	device.IsActive = false
	device.UpdatedAt = time.Now()

	return s.deviceRepo.Update(device)
}

// System operations
func (s *notificationService) ProcessScheduledNotifications() error {
	// Get notifications scheduled for now
	// This would typically involve a more complex query
	// For now, this is a placeholder implementation
	return nil
}

func (s *notificationService) CleanupExpiredNotifications() error {
	// Clean up expired notifications
	// This would involve deleting notifications past their expiry date
	// For now, this is a placeholder implementation
	return nil
}

// Helper methods
func (s *notificationService) processNotification(notification *domain.Notification) {
	var err error

	switch notification.Channel {
	case domain.ChannelSMS:
		err = s.sendSMS(notification)
	case domain.ChannelEmail:
		err = s.sendEmail(notification)
	case domain.ChannelPush:
		err = s.sendPush(notification)
	case domain.ChannelInApp:
		// In-app notifications are just stored in database
		err = nil
	}

	// Update notification status
	if err != nil {
		notification.Status = domain.StatusFailed
	} else {
		notification.Status = domain.StatusSent
		now := time.Now()
		notification.SentAt = &now
	}

	notification.UpdatedAt = time.Now()
	s.notificationRepo.Update(notification)
}

func (s *notificationService) sendSMS(notification *domain.Notification) error {
	// Get user's phone number from data
	phone, ok := notification.Data["phone"]
	if !ok {
		return errors.New("phone number not provided")
	}

	return s.smsService.SendSMS(phone, notification.Message)
}

func (s *notificationService) sendEmail(notification *domain.Notification) error {
	// Get user's email from data
	email, ok := notification.Data["email"]
	if !ok {
		return errors.New("email not provided")
	}

	return s.emailService.SendEmail(email, notification.Title, notification.Message)
}

func (s *notificationService) sendPush(notification *domain.Notification) error {
	// Get user's devices
	devices, err := s.deviceRepo.GetByUserID(notification.UserID)
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		return errors.New("no registered devices found")
	}

	// Collect active device tokens
	var deviceTokens []string
	for _, device := range devices {
		if device.IsActive {
			deviceTokens = append(deviceTokens, device.DeviceToken)
		}
	}

	if len(deviceTokens) == 0 {
		return errors.New("no active devices found")
	}

	return s.pushService.SendPushNotification(deviceTokens, notification.Title, notification.Message, notification.Data)
}

func (s *notificationService) replaceVariables(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = fmt.Sprintf(result, placeholder, value)
	}
	return result
}

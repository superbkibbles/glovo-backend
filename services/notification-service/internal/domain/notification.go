package domain

import (
	"time"
)

// Notification represents a notification entity
type Notification struct {
	ID           string               `json:"id" gorm:"primaryKey"`
	UserID       string               `json:"user_id" gorm:"index"`
	Type         NotificationType     `json:"type"`
	Channel      NotificationChannel  `json:"channel"`
	Title        string               `json:"title"`
	Message      string               `json:"message"`
	Data         map[string]string    `json:"data" gorm:"serializer:json"`
	Status       NotificationStatus   `json:"status"`
	Priority     NotificationPriority `json:"priority"`
	ScheduledFor *time.Time           `json:"scheduled_for,omitempty"`
	SentAt       *time.Time           `json:"sent_at,omitempty"`
	ReadAt       *time.Time           `json:"read_at,omitempty"`
	ExpiresAt    *time.Time           `json:"expires_at,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

type NotificationType string

const (
	TypeOrderUpdate      NotificationType = "order_update"
	TypeOrderConfirmed   NotificationType = "order_confirmed"
	TypeOrderDelivered   NotificationType = "order_delivered"
	TypeOrderCancelled   NotificationType = "order_cancelled"
	TypeDeliveryAssigned NotificationType = "delivery_assigned"
	TypeDeliveryUpdate   NotificationType = "delivery_update"
	TypePaymentSuccess   NotificationType = "payment_success"
	TypePaymentFailed    NotificationType = "payment_failed"
	TypePromotion        NotificationType = "promotion"
	TypeSystemAlert      NotificationType = "system_alert"
	TypeOTP              NotificationType = "otp"
	TypeWelcome          NotificationType = "welcome"
	TypeReminder         NotificationType = "reminder"
)

type NotificationChannel string

const (
	ChannelPush  NotificationChannel = "push"
	ChannelSMS   NotificationChannel = "sms"
	ChannelEmail NotificationChannel = "email"
	ChannelInApp NotificationChannel = "in_app"
)

type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusSent      NotificationStatus = "sent"
	StatusDelivered NotificationStatus = "delivered"
	StatusFailed    NotificationStatus = "failed"
	StatusRead      NotificationStatus = "read"
	StatusExpired   NotificationStatus = "expired"
)

type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "low"
	PriorityNormal   NotificationPriority = "normal"
	PriorityHigh     NotificationPriority = "high"
	PriorityCritical NotificationPriority = "critical"
)

// NotificationTemplate represents reusable notification templates
type NotificationTemplate struct {
	ID        string              `json:"id" gorm:"primaryKey"`
	Name      string              `json:"name" gorm:"uniqueIndex"`
	Type      NotificationType    `json:"type"`
	Channel   NotificationChannel `json:"channel"`
	Title     string              `json:"title"`
	Message   string              `json:"message"`
	Variables []string            `json:"variables" gorm:"serializer:json"`
	IsActive  bool                `json:"is_active"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// UserPreference represents user notification preferences
type UserPreference struct {
	ID        string              `json:"id" gorm:"primaryKey"`
	UserID    string              `json:"user_id" gorm:"index"`
	Type      NotificationType    `json:"type"`
	Channel   NotificationChannel `json:"channel"`
	Enabled   bool                `json:"enabled"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// NotificationDevice represents user devices for push notifications
type NotificationDevice struct {
	ID           string         `json:"id" gorm:"primaryKey"`
	UserID       string         `json:"user_id" gorm:"index"`
	DeviceToken  string         `json:"device_token" gorm:"uniqueIndex"`
	Platform     DevicePlatform `json:"platform"`
	AppVersion   string         `json:"app_version"`
	IsActive     bool           `json:"is_active"`
	LastActiveAt time.Time      `json:"last_active_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type DevicePlatform string

const (
	PlatformIOS     DevicePlatform = "ios"
	PlatformAndroid DevicePlatform = "android"
	PlatformWeb     DevicePlatform = "web"
)

// Request/Response DTOs
type SendNotificationRequest struct {
	UserID       string               `json:"user_id" binding:"required"`
	Type         NotificationType     `json:"type" binding:"required"`
	Channel      NotificationChannel  `json:"channel" binding:"required"`
	Title        string               `json:"title" binding:"required"`
	Message      string               `json:"message" binding:"required"`
	Data         map[string]string    `json:"data,omitempty"`
	Priority     NotificationPriority `json:"priority"`
	ScheduledFor *time.Time           `json:"scheduled_for,omitempty"`
	ExpiresAt    *time.Time           `json:"expires_at,omitempty"`
}

type BulkNotificationRequest struct {
	UserIDs      []string             `json:"user_ids" binding:"required"`
	Type         NotificationType     `json:"type" binding:"required"`
	Channel      NotificationChannel  `json:"channel" binding:"required"`
	Title        string               `json:"title" binding:"required"`
	Message      string               `json:"message" binding:"required"`
	Data         map[string]string    `json:"data,omitempty"`
	Priority     NotificationPriority `json:"priority"`
	ScheduledFor *time.Time           `json:"scheduled_for,omitempty"`
}

type SendTemplateNotificationRequest struct {
	UserID       string            `json:"user_id" binding:"required"`
	TemplateID   string            `json:"template_id" binding:"required"`
	Variables    map[string]string `json:"variables,omitempty"`
	ScheduledFor *time.Time        `json:"scheduled_for,omitempty"`
}

type UpdatePreferenceRequest struct {
	Type    NotificationType    `json:"type" binding:"required"`
	Channel NotificationChannel `json:"channel" binding:"required"`
	Enabled bool                `json:"enabled"`
}

type RegisterDeviceRequest struct {
	UserID      string         `json:"user_id" binding:"required"`
	DeviceToken string         `json:"device_token" binding:"required"`
	Platform    DevicePlatform `json:"platform" binding:"required"`
	AppVersion  string         `json:"app_version"`
}

type NotificationListResponse struct {
	Notifications []Notification `json:"notifications"`
	UnreadCount   int            `json:"unread_count"`
	TotalCount    int            `json:"total_count"`
}

// Repository interfaces (ports)
type NotificationRepository interface {
	Create(notification *Notification) error
	GetByID(id string) (*Notification, error)
	GetByUserID(userID string, limit, offset int) ([]Notification, error)
	GetUnreadByUserID(userID string) ([]Notification, error)
	GetByStatus(status NotificationStatus, limit, offset int) ([]Notification, error)
	Update(notification *Notification) error
	Delete(id string) error
	MarkAsRead(id string) error
	MarkAllAsRead(userID string) error
	GetUnreadCount(userID string) (int, error)
}

type TemplateRepository interface {
	Create(template *NotificationTemplate) error
	GetByID(id string) (*NotificationTemplate, error)
	GetByName(name string) (*NotificationTemplate, error)
	GetByType(notificationType NotificationType) ([]NotificationTemplate, error)
	Update(template *NotificationTemplate) error
	Delete(id string) error
	List(limit, offset int) ([]NotificationTemplate, error)
}

type PreferenceRepository interface {
	Create(preference *UserPreference) error
	GetByUserID(userID string) ([]UserPreference, error)
	GetByUserAndType(userID string, notificationType NotificationType) (*UserPreference, error)
	Update(preference *UserPreference) error
	Delete(id string) error
	UpsertPreference(userID string, notificationType NotificationType, channel NotificationChannel, enabled bool) error
}

type DeviceRepository interface {
	Create(device *NotificationDevice) error
	GetByID(id string) (*NotificationDevice, error)
	GetByUserID(userID string) ([]NotificationDevice, error)
	GetByToken(deviceToken string) (*NotificationDevice, error)
	Update(device *NotificationDevice) error
	Delete(id string) error
	DeactivateDevice(deviceToken string) error
	UpdateLastActive(deviceToken string) error
}

// Service interfaces (ports)
type NotificationService interface {
	// Sending notifications
	SendNotification(req SendNotificationRequest) (*Notification, error)
	SendBulkNotification(req BulkNotificationRequest) ([]Notification, error)
	SendTemplateNotification(req SendTemplateNotificationRequest) (*Notification, error)

	// OTP notifications (for User Service integration)
	SendOTPNotification(phoneNumber, otpCode string) error

	// Order notifications (for Order Service integration)
	SendOrderNotification(orderID, userID, message string) error

	// Managing notifications
	GetNotifications(userID string, limit, offset int) (*NotificationListResponse, error)
	GetNotification(notificationID string) (*Notification, error)
	MarkNotificationAsRead(notificationID, userID string) error
	MarkAllNotificationsAsRead(userID string) error
	GetUnreadCount(userID string) (int, error)

	// Templates
	CreateTemplate(template *NotificationTemplate) (*NotificationTemplate, error)
	GetTemplate(templateID string) (*NotificationTemplate, error)
	UpdateTemplate(templateID string, updates map[string]interface{}) (*NotificationTemplate, error)
	DeleteTemplate(templateID string) error
	ListTemplates(limit, offset int) ([]NotificationTemplate, error)

	// User preferences
	GetUserPreferences(userID string) ([]UserPreference, error)
	UpdateUserPreference(userID string, req UpdatePreferenceRequest) (*UserPreference, error)

	// Device management
	RegisterDevice(userID string, req RegisterDeviceRequest) (*NotificationDevice, error)
	GetUserDevices(userID string) ([]NotificationDevice, error)
	DeactivateDevice(userID, deviceToken string) error

	// System operations
	ProcessScheduledNotifications() error
	CleanupExpiredNotifications() error
}

// External service interfaces
type PushNotificationService interface {
	SendPushNotification(deviceTokens []string, title, message string, data map[string]string) error
	SendPushToTopic(topic, title, message string, data map[string]string) error
}

type SMSService interface {
	SendSMS(phoneNumber, message string) error
	SendBulkSMS(phoneNumbers []string, message string) error
}

type EmailService interface {
	SendEmail(to, subject, body string) error
	SendBulkEmail(recipients []string, subject, body string) error
	SendTemplateEmail(to, templateID string, variables map[string]string) error
}

// Event handling
type NotificationEvent struct {
	Type      string            `json:"type"`
	UserID    string            `json:"user_id"`
	Data      map[string]string `json:"data"`
	Timestamp time.Time         `json:"timestamp"`
}

type EventHandler interface {
	HandleOrderEvent(event NotificationEvent) error
	HandlePaymentEvent(event NotificationEvent) error
	HandleDeliveryEvent(event NotificationEvent) error
	HandleUserEvent(event NotificationEvent) error
}

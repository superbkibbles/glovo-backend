package db

import (
	"glovo-backend/services/notification-service/internal/domain"

	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) domain.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(notification *domain.Notification) error {
	return r.db.Create(notification).Error
}

func (r *notificationRepository) GetByID(id string) (*domain.Notification, error) {
	var notification domain.Notification
	err := r.db.Where("id = ?", id).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) GetByUserID(userID string, limit, offset int) ([]domain.Notification, error) {
	var notifications []domain.Notification
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) GetUnreadByUserID(userID string) ([]domain.Notification, error) {
	var notifications []domain.Notification
	err := r.db.Where("user_id = ? AND status != ?", userID, domain.StatusRead).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) GetByStatus(status domain.NotificationStatus, limit, offset int) ([]domain.Notification, error) {
	var notifications []domain.Notification
	err := r.db.Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) Update(notification *domain.Notification) error {
	return r.db.Save(notification).Error
}

func (r *notificationRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.Notification{}).Error
}

func (r *notificationRepository) MarkAsRead(id string) error {
	return r.db.Model(&domain.Notification{}).
		Where("id = ?", id).
		Update("status", domain.StatusRead).Error
}

func (r *notificationRepository) MarkAllAsRead(userID string) error {
	return r.db.Model(&domain.Notification{}).
		Where("user_id = ? AND status != ?", userID, domain.StatusRead).
		Update("status", domain.StatusRead).Error
}

func (r *notificationRepository) GetUnreadCount(userID string) (int, error) {
	var count int64
	err := r.db.Model(&domain.Notification{}).
		Where("user_id = ? AND status != ?", userID, domain.StatusRead).
		Count(&count).Error
	return int(count), err
}

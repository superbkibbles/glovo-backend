package db

import (
	"time"

	"glovo-backend/services/notification-service/internal/domain"

	"gorm.io/gorm"
)

type deviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) domain.DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) Create(device *domain.NotificationDevice) error {
	return r.db.Create(device).Error
}

func (r *deviceRepository) GetByID(id string) (*domain.NotificationDevice, error) {
	var device domain.NotificationDevice
	err := r.db.Where("id = ?", id).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) GetByUserID(userID string) ([]domain.NotificationDevice, error) {
	var devices []domain.NotificationDevice
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("last_active_at DESC").
		Find(&devices).Error
	return devices, err
}

func (r *deviceRepository) GetByToken(deviceToken string) (*domain.NotificationDevice, error) {
	var device domain.NotificationDevice
	err := r.db.Where("device_token = ?", deviceToken).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) Update(device *domain.NotificationDevice) error {
	return r.db.Save(device).Error
}

func (r *deviceRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.NotificationDevice{}).Error
}

func (r *deviceRepository) DeactivateDevice(deviceToken string) error {
	return r.db.Model(&domain.NotificationDevice{}).
		Where("device_token = ?", deviceToken).
		Update("is_active", false).Error
}

func (r *deviceRepository) UpdateLastActive(deviceToken string) error {
	return r.db.Model(&domain.NotificationDevice{}).
		Where("device_token = ?", deviceToken).
		Update("last_active_at", time.Now()).Error
}

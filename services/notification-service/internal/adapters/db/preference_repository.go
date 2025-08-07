package db

import (
	"glovo-backend/services/notification-service/internal/domain"

	"gorm.io/gorm"
)

type preferenceRepository struct {
	db *gorm.DB
}

func NewPreferenceRepository(db *gorm.DB) domain.PreferenceRepository {
	return &preferenceRepository{db: db}
}

func (r *preferenceRepository) Create(preference *domain.UserPreference) error {
	return r.db.Create(preference).Error
}

func (r *preferenceRepository) GetByUserID(userID string) ([]domain.UserPreference, error) {
	var preferences []domain.UserPreference
	err := r.db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&preferences).Error
	return preferences, err
}

func (r *preferenceRepository) GetByUserAndType(userID string, notificationType domain.NotificationType) (*domain.UserPreference, error) {
	var preference domain.UserPreference
	err := r.db.Where("user_id = ? AND type = ?", userID, notificationType).First(&preference).Error
	if err != nil {
		return nil, err
	}
	return &preference, nil
}

func (r *preferenceRepository) Update(preference *domain.UserPreference) error {
	return r.db.Save(preference).Error
}

func (r *preferenceRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.UserPreference{}).Error
}

func (r *preferenceRepository) UpsertPreference(userID string, notificationType domain.NotificationType, channel domain.NotificationChannel, enabled bool) error {
	preference := &domain.UserPreference{
		UserID:  userID,
		Type:    notificationType,
		Channel: channel,
		Enabled: enabled,
	}

	return r.db.Where(domain.UserPreference{UserID: userID, Type: notificationType, Channel: channel}).
		Assign(preference).
		FirstOrCreate(preference).Error
}

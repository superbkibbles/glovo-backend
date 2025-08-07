package db

import (
	"glovo-backend/services/notification-service/internal/domain"

	"gorm.io/gorm"
)

type templateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) domain.TemplateRepository {
	return &templateRepository{db: db}
}

func (r *templateRepository) Create(template *domain.NotificationTemplate) error {
	return r.db.Create(template).Error
}

func (r *templateRepository) GetByID(id string) (*domain.NotificationTemplate, error) {
	var template domain.NotificationTemplate
	err := r.db.Where("id = ?", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *templateRepository) GetByName(name string) (*domain.NotificationTemplate, error) {
	var template domain.NotificationTemplate
	err := r.db.Where("name = ?", name).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *templateRepository) GetByType(notificationType domain.NotificationType) ([]domain.NotificationTemplate, error) {
	var templates []domain.NotificationTemplate
	err := r.db.Where("type = ? AND is_active = ?", notificationType, true).
		Order("created_at DESC").
		Find(&templates).Error
	return templates, err
}

func (r *templateRepository) Update(template *domain.NotificationTemplate) error {
	return r.db.Save(template).Error
}

func (r *templateRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.NotificationTemplate{}).Error
}

func (r *templateRepository) List(limit, offset int) ([]domain.NotificationTemplate, error) {
	var templates []domain.NotificationTemplate
	err := r.db.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&templates).Error
	return templates, err
}

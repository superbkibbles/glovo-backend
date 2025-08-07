package db

import (
	"glovo-backend/services/admin-service/internal/domain"

	"gorm.io/gorm"
)

type systemConfigRepository struct {
	db *gorm.DB
}

func NewSystemConfigRepository(db *gorm.DB) domain.SystemConfigRepository {
	return &systemConfigRepository{db: db}
}

func (r *systemConfigRepository) Set(config *domain.SystemConfig) error {
	return r.db.Save(config).Error
}

func (r *systemConfigRepository) Get(key string) (*domain.SystemConfig, error) {
	var config domain.SystemConfig
	err := r.db.Where("key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *systemConfigRepository) GetAll() ([]domain.SystemConfig, error) {
	var configs []domain.SystemConfig
	err := r.db.Find(&configs).Error
	return configs, err
}

func (r *systemConfigRepository) Delete(key string) error {
	return r.db.Where("key = ?", key).Delete(&domain.SystemConfig{}).Error
}

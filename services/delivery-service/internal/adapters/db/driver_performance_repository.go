package db

import (
	"glovo-backend/services/delivery-service/internal/domain"

	"gorm.io/gorm"
)

type driverPerformanceRepository struct {
	db *gorm.DB
}

func NewDriverPerformanceRepository(db *gorm.DB) domain.DriverPerformanceRepository {
	return &driverPerformanceRepository{db: db}
}

func (r *driverPerformanceRepository) Create(performance *domain.DriverPerformance) error {
	return r.db.Create(performance).Error
}

func (r *driverPerformanceRepository) GetByDriverID(driverID string) (*domain.DriverPerformance, error) {
	var performance domain.DriverPerformance
	err := r.db.Where("driver_id = ?", driverID).First(&performance).Error
	if err != nil {
		return nil, err
	}
	return &performance, nil
}

func (r *driverPerformanceRepository) Update(performance *domain.DriverPerformance) error {
	return r.db.Save(performance).Error
}

func (r *driverPerformanceRepository) GetTopDrivers(limit int) ([]domain.DriverPerformance, error) {
	var performances []domain.DriverPerformance
	err := r.db.Order("average_rating DESC, completed_deliveries DESC").
		Limit(limit).
		Find(&performances).Error
	return performances, err
}

func (r *driverPerformanceRepository) GetDriverRankings() ([]domain.DriverPerformance, error) {
	var performances []domain.DriverPerformance
	err := r.db.Order("average_rating DESC, completed_deliveries DESC").
		Find(&performances).Error
	return performances, err
}

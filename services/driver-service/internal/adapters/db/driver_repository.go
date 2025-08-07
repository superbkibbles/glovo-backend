package db

import (
	"glovo-backend/services/driver-service/internal/domain"

	"gorm.io/gorm"
)

type driverRepository struct {
	db *gorm.DB
}

func NewDriverRepository(db *gorm.DB) domain.DriverRepository {
	return &driverRepository{db: db}
}

func (r *driverRepository) Create(driver *domain.Driver) error {
	return r.db.Create(driver).Error
}

func (r *driverRepository) GetByID(id string) (*domain.Driver, error) {
	var driver domain.Driver
	err := r.db.Preload("Documents").Where("id = ?", id).First(&driver).Error
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

func (r *driverRepository) GetByUserID(userID string) (*domain.Driver, error) {
	var driver domain.Driver
	err := r.db.Preload("Documents").Where("user_id = ?", userID).First(&driver).Error
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

func (r *driverRepository) Update(driver *domain.Driver) error {
	return r.db.Save(driver).Error
}

func (r *driverRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.Driver{}).Error
}

func (r *driverRepository) Search(req domain.DriverSearchRequest) ([]domain.Driver, error) {
	query := r.db.Model(&domain.Driver{}).Preload("Documents")

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.VehicleType != "" {
		query = query.Where("vehicle_type = ?", req.VehicleType)
	}

	if req.MinRating > 0 {
		query = query.Where("performance_rating >= ?", req.MinRating)
	}

	if req.Available != nil {
		query = query.Where("availability_is_available = ?", *req.Available)
	}

	// Location-based search
	if req.Latitude != 0 && req.Longitude != 0 && req.Radius > 0 {
		// Simple distance calculation (in production, use proper geospatial queries)
		query = query.Where(
			"(6371 * acos(cos(radians(?)) * cos(radians(location_latitude)) * cos(radians(location_longitude) - radians(?)) + sin(radians(?)) * sin(radians(location_latitude)))) < ?",
			req.Latitude, req.Longitude, req.Latitude, req.Radius,
		)
	}

	if req.Limit == 0 {
		req.Limit = 20
	}

	var drivers []domain.Driver
	err := query.Order("performance_rating DESC").
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&drivers).Error

	return drivers, err
}

func (r *driverRepository) List(limit, offset int) ([]domain.Driver, error) {
	var drivers []domain.Driver
	err := r.db.Preload("Documents").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&drivers).Error
	return drivers, err
}

func (r *driverRepository) GetByStatus(status domain.DriverStatus, limit, offset int) ([]domain.Driver, error) {
	var drivers []domain.Driver
	err := r.db.Where("status = ?", status).
		Preload("Documents").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&drivers).Error
	return drivers, err
}

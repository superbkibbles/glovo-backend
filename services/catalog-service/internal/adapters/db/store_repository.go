package db

import (
	"fmt"
	"strings"

	"glovo-backend/services/catalog-service/internal/domain"

	"gorm.io/gorm"
)

type storeRepository struct {
	db *gorm.DB
}

func NewStoreRepository(db *gorm.DB) domain.StoreRepository {
	return &storeRepository{db: db}
}

func (r *storeRepository) Create(store *domain.Store) error {
	return r.db.Create(store).Error
}

func (r *storeRepository) GetByID(id string) (*domain.Store, error) {
	var store domain.Store
	err := r.db.Preload("Categories").Preload("Products").Where("id = ?", id).First(&store).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *storeRepository) GetByMerchantID(merchantID string) (*domain.Store, error) {
	var store domain.Store
	err := r.db.Preload("Categories").Preload("Products").Where("merchant_id = ?", merchantID).First(&store).Error
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *storeRepository) Update(store *domain.Store) error {
	return r.db.Save(store).Error
}

func (r *storeRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.Store{}).Error
}

func (r *storeRepository) Search(req domain.StoreSearchRequest) ([]domain.Store, error) {
	query := r.db.Model(&domain.Store{}).Preload("Categories")

	// Apply filters
	if req.Query != "" {
		searchTerm := "%" + strings.ToLower(req.Query) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	if req.CategoryID != "" {
		query = query.Joins("JOIN store_categories ON stores.id = store_categories.store_id").
			Where("store_categories.category_id = ?", req.CategoryID)
	}

	if req.MinRating > 0 {
		query = query.Where("rating >= ?", req.MinRating)
	}

	// Location-based filtering
	if req.Latitude != 0 && req.Longitude != 0 && req.Radius > 0 {
		// Using Haversine formula for distance calculation
		// This is a simplified version; in production, you might want to use PostGIS
		query = query.Where(
			`(6371 * acos(cos(radians(?)) * cos(radians(latitude)) * cos(radians(longitude) - radians(?)) + sin(radians(?)) * sin(radians(latitude)))) <= ?`,
			req.Latitude, req.Longitude, req.Latitude, req.Radius,
		)
	}

	// Sorting
	switch req.SortBy {
	case "rating":
		query = query.Order("rating DESC")
	case "distance":
		if req.Latitude != 0 && req.Longitude != 0 {
			query = query.Order(fmt.Sprintf(
				"(6371 * acos(cos(radians(%f)) * cos(radians(latitude)) * cos(radians(longitude) - radians(%f)) + sin(radians(%f)) * sin(radians(latitude)))) ASC",
				req.Latitude, req.Longitude, req.Latitude,
			))
		}
	case "delivery_time":
		query = query.Order("delivery_info_estimated_time ASC")
	default:
		query = query.Order("created_at DESC")
	}

	// Pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	} else {
		query = query.Limit(20) // Default limit
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	var stores []domain.Store
	err := query.Find(&stores).Error
	return stores, err
}

func (r *storeRepository) List(limit, offset int) ([]domain.Store, error) {
	var stores []domain.Store
	err := r.db.Preload("Categories").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&stores).Error
	return stores, err
}

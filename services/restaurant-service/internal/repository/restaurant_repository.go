package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mendmzury/food-delivery/services/restaurant-service/internal/domain"
	"gorm.io/gorm"
)

type RestaurantRepository struct {
	db *gorm.DB
}

func NewRestaurantRepository(db *gorm.DB) *RestaurantRepository {
	return &RestaurantRepository{db: db}
}

func (r *RestaurantRepository) Create(ctx context.Context, restaurant *domain.Restaurant) error {
	if restaurant.ID == uuid.Nil {
		restaurant.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(restaurant).Error
}

func (r *RestaurantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error) {
	var restaurant domain.Restaurant
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&restaurant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &restaurant, nil
}

func (r *RestaurantRepository) List(ctx context.Context, search, status string, page, pageSize int64) ([]*domain.Restaurant, int64, error) {
	var restaurants []*domain.Restaurant
	query := r.db.WithContext(ctx).Model(&domain.Restaurant{})

	if search != "" {
		s := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(address) LIKE ?", s, s, s)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&restaurants).Error
	return restaurants, total, err
}

func (r *RestaurantRepository) Update(ctx context.Context, restaurant *domain.Restaurant) error {
	return r.db.WithContext(ctx).Save(restaurant).Error
}

func (r *RestaurantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Restaurant{}).Error
}

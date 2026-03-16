package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mendmzury/food-delivery/services/restaurant-service/internal/domain"
	"gorm.io/gorm"
)

type MenuItemRepository struct {
	db *gorm.DB
}

func NewMenuItemRepository(db *gorm.DB) *MenuItemRepository {
	return &MenuItemRepository{db: db}
}

func (r *MenuItemRepository) Create(ctx context.Context, item *domain.MenuItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *MenuItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.MenuItem, error) {
	var item domain.MenuItem
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *MenuItemRepository) ListByCategory(ctx context.Context, categoryID uuid.UUID, availableOnly bool) ([]*domain.MenuItem, error) {
	var items []*domain.MenuItem
	query := r.db.WithContext(ctx).Where("category_id = ?", categoryID)
	if availableOnly {
		query = query.Where("available = ?", true)
	}
	err := query.Order("sort_order ASC, created_at ASC").Find(&items).Error
	return items, err
}

func (r *MenuItemRepository) ListByRestaurant(ctx context.Context, restaurantID uuid.UUID, availableOnly bool) ([]*domain.MenuItem, error) {
	var categoryIDs []uuid.UUID
	if err := r.db.WithContext(ctx).Model(&domain.Category{}).Where("restaurant_id = ?", restaurantID).Pluck("id", &categoryIDs).Error; err != nil {
		return nil, err
	}
	if len(categoryIDs) == 0 {
		return []*domain.MenuItem{}, nil
	}
	query := r.db.WithContext(ctx).Where("category_id IN ?", categoryIDs)
	if availableOnly {
		query = query.Where("available = ?", true)
	}
	var items []*domain.MenuItem
	err := query.Order("sort_order ASC, created_at ASC").Find(&items).Error
	return items, err
}

func (r *MenuItemRepository) Update(ctx context.Context, item *domain.MenuItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *MenuItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.MenuItem{}).Error
}

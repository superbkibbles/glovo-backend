package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mendmzury/food-delivery/services/order-service/internal/domain"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *domain.Order, items []*domain.OrderItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if order.ID == uuid.Nil {
			order.ID = uuid.New()
		}
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for _, item := range items {
			if item.ID == uuid.Nil {
				item.ID = uuid.New()
			}
			item.OrderID = order.ID
			if err := tx.Create(item).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, []*domain.OrderItem, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	var items []*domain.OrderItem
	if err := r.db.WithContext(ctx).Where("order_id = ?", id).Find(&items).Error; err != nil {
		return nil, nil, err
	}
	return &order, items, nil
}

func (r *OrderRepository) List(ctx context.Context, customerID, restaurantID, driverID *uuid.UUID, status domain.OrderStatus, page, pageSize int64) ([]*domain.Order, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.Order{})

	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}
	if restaurantID != nil {
		query = query.Where("restaurant_id = ?", *restaurantID)
	}
	if driverID != nil {
		query = query.Where("driver_id = ?", *driverID)
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

	var orders []*domain.Order
	err := query.Order("created_at DESC").Offset(int(offset)).Limit(int(pageSize)).Find(&orders).Error
	return orders, total, err
}

func (r *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *OrderRepository) GetItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderItem, error) {
	var items []*domain.OrderItem
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}

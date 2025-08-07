package db

import (
	"glovo-backend/services/order-service/internal/domain"

	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) domain.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *domain.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) GetByID(id string) (*domain.Order, error) {
	var order domain.Order
	err := r.db.Preload("Items").Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByCustomerID(customerID string, limit, offset int) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.Preload("Items").
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetByMerchantID(merchantID string, limit, offset int) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.Preload("Items").
		Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetByDriverID(driverID string, limit, offset int) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.Preload("Items").
		Where("driver_id = ?", driverID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetByStatus(status domain.OrderStatus, limit, offset int) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.Preload("Items").
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) Update(order *domain.Order) error {
	return r.db.Save(order).Error
}

func (r *orderRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.Order{}).Error
}

func (r *orderRepository) List(limit, offset int) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.Preload("Items").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	return orders, err
}

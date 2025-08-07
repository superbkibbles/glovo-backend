package db

import (
	"glovo-backend/services/delivery-service/internal/domain"

	"gorm.io/gorm"
)

type deliveryRepository struct {
	db *gorm.DB
}

func NewDeliveryRepository(db *gorm.DB) domain.DeliveryRepository {
	return &deliveryRepository{db: db}
}

func (r *deliveryRepository) Create(delivery *domain.Delivery) error {
	return r.db.Create(delivery).Error
}

func (r *deliveryRepository) GetByID(id string) (*domain.Delivery, error) {
	var delivery domain.Delivery
	err := r.db.Where("id = ?", id).First(&delivery).Error
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *deliveryRepository) GetByOrderID(orderID string) (*domain.Delivery, error) {
	var delivery domain.Delivery
	err := r.db.Where("order_id = ?", orderID).First(&delivery).Error
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (r *deliveryRepository) GetByDriverID(driverID string, limit, offset int) ([]domain.Delivery, error) {
	var deliveries []domain.Delivery
	err := r.db.Where("driver_id = ?", driverID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&deliveries).Error
	return deliveries, err
}

func (r *deliveryRepository) GetByStatus(status domain.DeliveryStatus, limit, offset int) ([]domain.Delivery, error) {
	var deliveries []domain.Delivery
	err := r.db.Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&deliveries).Error
	return deliveries, err
}

func (r *deliveryRepository) Search(req domain.DeliverySearchRequest) ([]domain.Delivery, error) {
	query := r.db.Model(&domain.Delivery{})

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.DriverID != "" {
		query = query.Where("driver_id = ?", req.DriverID)
	}
	if req.Priority != "" {
		query = query.Where("priority = ?", req.Priority)
	}
	if req.DateFrom != nil {
		query = query.Where("created_at >= ?", req.DateFrom)
	}
	if req.DateTo != nil {
		query = query.Where("created_at <= ?", req.DateTo)
	}

	if req.Limit == 0 {
		req.Limit = 20
	}

	var deliveries []domain.Delivery
	err := query.Order("created_at DESC").
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&deliveries).Error
	return deliveries, err
}

func (r *deliveryRepository) Update(delivery *domain.Delivery) error {
	return r.db.Save(delivery).Error
}

func (r *deliveryRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.Delivery{}).Error
}

func (r *deliveryRepository) GetPendingDeliveries() ([]domain.Delivery, error) {
	var deliveries []domain.Delivery
	err := r.db.Where("status = ?", domain.StatusPending).Find(&deliveries).Error
	return deliveries, err
}

func (r *deliveryRepository) GetActiveDeliveries() ([]domain.Delivery, error) {
	var deliveries []domain.Delivery
	err := r.db.Where("status IN ?", []domain.DeliveryStatus{
		domain.StatusAssigned,
		domain.StatusAccepted,
		domain.StatusPickedUp,
		domain.StatusInTransit,
	}).Find(&deliveries).Error
	return deliveries, err
}

package db

import (
	"glovo-backend/services/payment-service/internal/domain"

	"gorm.io/gorm"
)

type commissionRepository struct {
	db *gorm.DB
}

func NewCommissionRepository(db *gorm.DB) domain.CommissionRepository {
	return &commissionRepository{db: db}
}

func (r *commissionRepository) Create(commission *domain.Commission) error {
	return r.db.Create(commission).Error
}

func (r *commissionRepository) GetByID(id string) (*domain.Commission, error) {
	var commission domain.Commission
	err := r.db.Where("id = ?", id).First(&commission).Error
	if err != nil {
		return nil, err
	}
	return &commission, nil
}

func (r *commissionRepository) GetByOrderID(orderID string) (*domain.Commission, error) {
	var commission domain.Commission
	err := r.db.Where("order_id = ?", orderID).First(&commission).Error
	if err != nil {
		return nil, err
	}
	return &commission, nil
}

func (r *commissionRepository) GetByMerchantID(merchantID string, limit, offset int) ([]domain.Commission, error) {
	var commissions []domain.Commission
	err := r.db.Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&commissions).Error
	return commissions, err
}

func (r *commissionRepository) GetByDriverID(driverID string, limit, offset int) ([]domain.Commission, error) {
	var commissions []domain.Commission
	err := r.db.Where("driver_id = ?", driverID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&commissions).Error
	return commissions, err
}

func (r *commissionRepository) Update(commission *domain.Commission) error {
	return r.db.Save(commission).Error
}

func (r *commissionRepository) List(limit, offset int) ([]domain.Commission, error) {
	var commissions []domain.Commission
	err := r.db.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&commissions).Error
	return commissions, err
}

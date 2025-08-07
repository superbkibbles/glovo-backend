package db

import (
	"glovo-backend/services/payment-service/internal/domain"

	"gorm.io/gorm"
)

type paymentMethodRepository struct {
	db *gorm.DB
}

func NewPaymentMethodRepository(db *gorm.DB) domain.PaymentMethodRepository {
	return &paymentMethodRepository{db: db}
}

func (r *paymentMethodRepository) Create(method *domain.PaymentMethod) error {
	return r.db.Create(method).Error
}

func (r *paymentMethodRepository) GetByID(id string) (*domain.PaymentMethod, error) {
	var method domain.PaymentMethod
	err := r.db.Where("id = ?", id).First(&method).Error
	if err != nil {
		return nil, err
	}
	return &method, nil
}

func (r *paymentMethodRepository) GetByUserID(userID string) ([]domain.PaymentMethod, error) {
	var methods []domain.PaymentMethod
	err := r.db.Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&methods).Error
	return methods, err
}

func (r *paymentMethodRepository) Update(method *domain.PaymentMethod) error {
	return r.db.Save(method).Error
}

func (r *paymentMethodRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.PaymentMethod{}).Error
}

func (r *paymentMethodRepository) SetAsDefault(userID, methodID string) error {
	// Start a transaction to ensure consistency
	return r.db.Transaction(func(tx *gorm.DB) error {
		// First, unset all existing default payment methods for the user
		if err := tx.Model(&domain.PaymentMethod{}).
			Where("user_id = ?", userID).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// Then set the specified method as default
		return tx.Model(&domain.PaymentMethod{}).
			Where("id = ? AND user_id = ?", methodID, userID).
			Update("is_default", true).Error
	})
}

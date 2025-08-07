package db

import (
	"glovo-backend/services/payment-service/internal/domain"

	"gorm.io/gorm"
)

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) domain.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) Create(wallet *domain.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *walletRepository) GetByID(id string) (*domain.Wallet, error) {
	var wallet domain.Wallet
	err := r.db.Where("id = ?", id).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) GetByUserID(userID string) (*domain.Wallet, error) {
	var wallet domain.Wallet
	err := r.db.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) Update(wallet *domain.Wallet) error {
	return r.db.Save(wallet).Error
}

func (r *walletRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&domain.Wallet{}).Error
}

func (r *walletRepository) List(limit, offset int) ([]domain.Wallet, error) {
	var wallets []domain.Wallet
	err := r.db.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&wallets).Error
	return wallets, err
}

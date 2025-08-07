package db

import (
	"time"

	"glovo-backend/services/payment-service/internal/domain"

	"gorm.io/gorm"
)

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) domain.TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(transaction *domain.Transaction) error {
	return r.db.Create(transaction).Error
}

func (r *transactionRepository) GetByID(id string) (*domain.Transaction, error) {
	var transaction domain.Transaction
	err := r.db.Where("id = ?", id).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) GetByWalletID(walletID string, limit, offset int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := r.db.Where("from_wallet_id = ? OR to_wallet_id = ?", walletID, walletID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) GetByUserID(userID string, limit, offset int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction

	// Join with wallets to find transactions for a specific user
	err := r.db.Table("transactions").
		Select("transactions.*").
		Joins("LEFT JOIN wallets w1 ON transactions.from_wallet_id = w1.id").
		Joins("LEFT JOIN wallets w2 ON transactions.to_wallet_id = w2.id").
		Where("w1.user_id = ? OR w2.user_id = ?", userID, userID).
		Order("transactions.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	return transactions, err
}

func (r *transactionRepository) GetByOrderID(orderID string) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := r.db.Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&transactions).Error
	return transactions, err
}

func (r *transactionRepository) Update(transaction *domain.Transaction) error {
	return r.db.Save(transaction).Error
}

func (r *transactionRepository) GetTransactionReport(userID string, startDate, endDate time.Time) (*domain.TransactionReport, error) {
	var result struct {
		TotalAmount      float64 `gorm:"column:total_amount"`
		TotalFees        float64 `gorm:"column:total_fees"`
		TransactionCount int     `gorm:"column:transaction_count"`
	}

	err := r.db.Table("transactions").
		Select("COALESCE(SUM(amount), 0) as total_amount, COALESCE(SUM(fee), 0) as total_fees, COUNT(*) as transaction_count").
		Joins("LEFT JOIN wallets w1 ON transactions.from_wallet_id = w1.id").
		Joins("LEFT JOIN wallets w2 ON transactions.to_wallet_id = w2.id").
		Where("(w1.user_id = ? OR w2.user_id = ?) AND transactions.created_at BETWEEN ? AND ?", userID, userID, startDate, endDate).
		Where("transactions.status = ?", domain.TxStatusCompleted).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	report := &domain.TransactionReport{
		Period:           startDate.Format("2006-01-02") + " to " + endDate.Format("2006-01-02"),
		TotalAmount:      result.TotalAmount,
		TotalFees:        result.TotalFees,
		TransactionCount: result.TransactionCount,
		NetAmount:        result.TotalAmount - result.TotalFees,
	}

	return report, nil
}

func (r *transactionRepository) List(limit, offset int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := r.db.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

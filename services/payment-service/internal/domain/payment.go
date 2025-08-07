package domain

import (
	"time"

	"glovo-backend/shared/auth"
)

// Wallet represents user wallet/account
type Wallet struct {
	ID             string        `json:"id" gorm:"primaryKey"`
	UserID         string        `json:"user_id" gorm:"uniqueIndex"`
	UserType       auth.UserRole `json:"user_type"`
	Balance        float64       `json:"balance"`
	PendingBalance float64       `json:"pending_balance"`
	Currency       string        `json:"currency"`
	Status         WalletStatus  `json:"status"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

type WalletStatus string

const (
	WalletStatusActive    WalletStatus = "active"
	WalletStatusSuspended WalletStatus = "suspended"
	WalletStatusFrozen    WalletStatus = "frozen"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID              string            `json:"id" gorm:"primaryKey"`
	FromWalletID    *string           `json:"from_wallet_id,omitempty"`
	ToWalletID      *string           `json:"to_wallet_id,omitempty"`
	Type            TransactionType   `json:"type"`
	Amount          float64           `json:"amount"`
	Fee             float64           `json:"fee"`
	NetAmount       float64           `json:"net_amount"`
	Currency        string            `json:"currency"`
	Status          TransactionStatus `json:"status"`
	Description     string            `json:"description"`
	Reference       string            `json:"reference,omitempty"`
	OrderID         *string           `json:"order_id,omitempty"`
	PaymentMethodID *string           `json:"payment_method_id,omitempty"`
	Metadata        map[string]string `json:"metadata" gorm:"serializer:json"`
	ProcessedAt     *time.Time        `json:"processed_at,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type TransactionType string

const (
	TxTypePayment    TransactionType = "payment"
	TxTypeRefund     TransactionType = "refund"
	TxTypeTransfer   TransactionType = "transfer"
	TxTypeTopUp      TransactionType = "top_up"
	TxTypeWithdrawal TransactionType = "withdrawal"
	TxTypeCommission TransactionType = "commission"
	TxTypePayout     TransactionType = "payout"
	TxTypeBonus      TransactionType = "bonus"
	TxTypePenalty    TransactionType = "penalty"
)

type TransactionStatus string

const (
	TxStatusPending   TransactionStatus = "pending"
	TxStatusCompleted TransactionStatus = "completed"
	TxStatusFailed    TransactionStatus = "failed"
	TxStatusCancelled TransactionStatus = "cancelled"
	TxStatusRefunded  TransactionStatus = "refunded"
)

// PaymentMethod represents user payment methods
type PaymentMethod struct {
	ID            string              `json:"id" gorm:"primaryKey"`
	UserID        string              `json:"user_id" gorm:"index"`
	Type          PaymentMethodType   `json:"type"`
	Provider      string              `json:"provider"`
	IsDefault     bool                `json:"is_default"`
	Status        PaymentMethodStatus `json:"status"`
	CardInfo      *CardInfo           `json:"card_info,omitempty" gorm:"embedded"`
	BankInfo      *BankAccountInfo    `json:"bank_info,omitempty" gorm:"embedded"`
	DigitalWallet *DigitalWalletInfo  `json:"digital_wallet,omitempty" gorm:"embedded"`
	ExpiresAt     *time.Time          `json:"expires_at,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

type PaymentMethodType string

const (
	PaymentTypeCard          PaymentMethodType = "card"
	PaymentTypeBankAccount   PaymentMethodType = "bank_account"
	PaymentTypeDigitalWallet PaymentMethodType = "digital_wallet"
	PaymentTypeCash          PaymentMethodType = "cash"
)

type PaymentMethodStatus string

const (
	PaymentStatusActive   PaymentMethodStatus = "active"
	PaymentStatusInactive PaymentMethodStatus = "inactive"
	PaymentStatusExpired  PaymentMethodStatus = "expired"
)

type CardInfo struct {
	LastFourDigits string `json:"last_four_digits"`
	Brand          string `json:"brand"`
	ExpiryMonth    int    `json:"expiry_month"`
	ExpiryYear     int    `json:"expiry_year"`
	HolderName     string `json:"holder_name"`
}

type BankAccountInfo struct {
	AccountNumber string `json:"account_number"`
	RoutingNumber string `json:"routing_number"`
	BankName      string `json:"bank_name"`
	AccountType   string `json:"account_type"`
	HolderName    string `json:"holder_name"`
}

type DigitalWalletInfo struct {
	WalletType string `json:"wallet_type"` // paypal, apple_pay, google_pay
	AccountID  string `json:"account_id"`
	Email      string `json:"email,omitempty"`
}

// Commission represents commission calculations
type Commission struct {
	ID            string           `json:"id" gorm:"primaryKey"`
	OrderID       string           `json:"order_id" gorm:"index"`
	MerchantID    string           `json:"merchant_id"`
	DriverID      *string          `json:"driver_id,omitempty"`
	OrderAmount   float64          `json:"order_amount"`
	PlatformFee   float64          `json:"platform_fee"`
	DeliveryFee   float64          `json:"delivery_fee"`
	DriverFee     float64          `json:"driver_fee"`
	MerchantFee   float64          `json:"merchant_fee"`
	NetToMerchant float64          `json:"net_to_merchant"`
	NetToDriver   float64          `json:"net_to_driver"`
	Status        CommissionStatus `json:"status"`
	ProcessedAt   *time.Time       `json:"processed_at,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
}

type CommissionStatus string

const (
	CommissionStatusPending   CommissionStatus = "pending"
	CommissionStatusProcessed CommissionStatus = "processed"
	CommissionStatusFailed    CommissionStatus = "failed"
)

// Request/Response DTOs
type ProcessPaymentRequest struct {
	OrderID         string            `json:"order_id" binding:"required"`
	CustomerID      string            `json:"customer_id" binding:"required"`
	Amount          float64           `json:"amount" binding:"required,min=0"`
	Currency        string            `json:"currency"`
	PaymentMethodID string            `json:"payment_method_id" binding:"required"`
	Description     string            `json:"description"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type RefundRequest struct {
	TransactionID string  `json:"transaction_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,min=0"`
	Reason        string  `json:"reason" binding:"required"`
}

type TransferRequest struct {
	FromUserID  string            `json:"from_user_id" binding:"required"`
	ToUserID    string            `json:"to_user_id" binding:"required"`
	Amount      float64           `json:"amount" binding:"required,min=0"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type TopUpRequest struct {
	UserID          string  `json:"user_id" binding:"required"`
	Amount          float64 `json:"amount" binding:"required,min=0"`
	PaymentMethodID string  `json:"payment_method_id" binding:"required"`
}

type WithdrawalRequest struct {
	UserID          string  `json:"user_id" binding:"required"`
	Amount          float64 `json:"amount" binding:"required,min=0"`
	PaymentMethodID string  `json:"payment_method_id" binding:"required"`
}

type AddPaymentMethodRequest struct {
	Type          PaymentMethodType  `json:"type" binding:"required"`
	Provider      string             `json:"provider" binding:"required"`
	CardToken     string             `json:"card_token,omitempty"`
	BankInfo      *BankAccountInfo   `json:"bank_info,omitempty"`
	DigitalWallet *DigitalWalletInfo `json:"digital_wallet,omitempty"`
	IsDefault     bool               `json:"is_default"`
}

type PaymentResponse struct {
	TransactionID string            `json:"transaction_id"`
	Status        TransactionStatus `json:"status"`
	Amount        float64           `json:"amount"`
	Fee           float64           `json:"fee"`
	NetAmount     float64           `json:"net_amount"`
	Reference     string            `json:"reference,omitempty"`
	ProcessedAt   *time.Time        `json:"processed_at,omitempty"`
}

type WalletBalance struct {
	Balance        float64 `json:"balance"`
	PendingBalance float64 `json:"pending_balance"`
	Currency       string  `json:"currency"`
}

type TransactionReport struct {
	Period           string  `json:"period"`
	TotalAmount      float64 `json:"total_amount"`
	TotalFees        float64 `json:"total_fees"`
	TransactionCount int     `json:"transaction_count"`
	NetAmount        float64 `json:"net_amount"`
}

// Repository interfaces (ports)
type WalletRepository interface {
	Create(wallet *Wallet) error
	GetByID(id string) (*Wallet, error)
	GetByUserID(userID string) (*Wallet, error)
	Update(wallet *Wallet) error
	Delete(id string) error
	List(limit, offset int) ([]Wallet, error)
}

type TransactionRepository interface {
	Create(transaction *Transaction) error
	GetByID(id string) (*Transaction, error)
	GetByWalletID(walletID string, limit, offset int) ([]Transaction, error)
	GetByUserID(userID string, limit, offset int) ([]Transaction, error)
	GetByOrderID(orderID string) ([]Transaction, error)
	Update(transaction *Transaction) error
	GetTransactionReport(userID string, startDate, endDate time.Time) (*TransactionReport, error)
	List(limit, offset int) ([]Transaction, error)
}

type PaymentMethodRepository interface {
	Create(method *PaymentMethod) error
	GetByID(id string) (*PaymentMethod, error)
	GetByUserID(userID string) ([]PaymentMethod, error)
	Update(method *PaymentMethod) error
	Delete(id string) error
	SetAsDefault(userID, methodID string) error
}

type CommissionRepository interface {
	Create(commission *Commission) error
	GetByID(id string) (*Commission, error)
	GetByOrderID(orderID string) (*Commission, error)
	GetByMerchantID(merchantID string, limit, offset int) ([]Commission, error)
	GetByDriverID(driverID string, limit, offset int) ([]Commission, error)
	Update(commission *Commission) error
	List(limit, offset int) ([]Commission, error)
}

// Service interfaces (ports)
type PaymentService interface {
	// Wallet management
	CreateWallet(userID string, userType auth.UserRole) (*Wallet, error)
	GetWallet(userID string) (*Wallet, error)
	GetBalance(userID string) (*WalletBalance, error)

	// Payment processing
	ProcessPayment(req ProcessPaymentRequest) (*PaymentResponse, error)
	ProcessRefund(req RefundRequest) (*PaymentResponse, error)
	ProcessTransfer(req TransferRequest) (*PaymentResponse, error)
	ProcessTopUp(req TopUpRequest) (*PaymentResponse, error)
	ProcessWithdrawal(req WithdrawalRequest) (*PaymentResponse, error)

	// Payment methods
	AddPaymentMethod(userID string, req AddPaymentMethodRequest) (*PaymentMethod, error)
	GetPaymentMethods(userID string) ([]PaymentMethod, error)
	SetDefaultPaymentMethod(userID, methodID string) error
	RemovePaymentMethod(userID, methodID string) error

	// Transactions
	GetTransaction(transactionID string) (*Transaction, error)
	GetTransactionHistory(userID string, limit, offset int) ([]Transaction, error)
	GetTransactionReport(userID string, startDate, endDate time.Time) (*TransactionReport, error)

	// Commission management
	CalculateCommission(orderID string, orderAmount, deliveryFee float64, merchantID, driverID string) (*Commission, error)
	ProcessCommission(commissionID string) error
	GetMerchantCommissions(merchantID string, limit, offset int) ([]Commission, error)
	GetDriverCommissions(driverID string, limit, offset int) ([]Commission, error)

	// Payouts
	ProcessMerchantPayout(merchantID string, amount float64) (*PaymentResponse, error)
	ProcessDriverPayout(driverID string, amount float64) (*PaymentResponse, error)
}

// External service interfaces
type StripeService interface {
	ProcessCardPayment(token string, amount float64, currency string) (*StripePaymentResult, error)
	CreateCustomer(userID, email string) (*StripeCustomer, error)
	RefundPayment(chargeID string, amount float64) (*StripeRefund, error)
}

type BankService interface {
	ProcessACHTransfer(accountInfo BankAccountInfo, amount float64) (*BankTransferResult, error)
	ValidateBankAccount(accountInfo BankAccountInfo) error
}

// External DTOs
type StripePaymentResult struct {
	ChargeID  string `json:"charge_id"`
	Status    string `json:"status"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	Reference string `json:"reference"`
}

type StripeCustomer struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type StripeRefund struct {
	ID       string `json:"id"`
	Amount   int64  `json:"amount"`
	Status   string `json:"status"`
	ChargeID string `json:"charge_id"`
}

type BankTransferResult struct {
	TransferID string  `json:"transfer_id"`
	Status     string  `json:"status"`
	Amount     float64 `json:"amount"`
	Reference  string  `json:"reference"`
}

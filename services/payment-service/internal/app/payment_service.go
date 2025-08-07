package app

import (
	"errors"
	"fmt"
	"time"

	"glovo-backend/services/payment-service/internal/domain"
	"glovo-backend/shared/auth"

	"github.com/google/uuid"
)

type paymentService struct {
	walletRepo        domain.WalletRepository
	transactionRepo   domain.TransactionRepository
	paymentMethodRepo domain.PaymentMethodRepository
	commissionRepo    domain.CommissionRepository
	stripeService     domain.StripeService
	bankService       domain.BankService
}

func NewPaymentService(
	walletRepo domain.WalletRepository,
	transactionRepo domain.TransactionRepository,
	paymentMethodRepo domain.PaymentMethodRepository,
	commissionRepo domain.CommissionRepository,
	stripeService domain.StripeService,
	bankService domain.BankService,
) domain.PaymentService {
	return &paymentService{
		walletRepo:        walletRepo,
		transactionRepo:   transactionRepo,
		paymentMethodRepo: paymentMethodRepo,
		commissionRepo:    commissionRepo,
		stripeService:     stripeService,
		bankService:       bankService,
	}
}

// Wallet management
func (s *paymentService) CreateWallet(userID string, userType auth.UserRole) (*domain.Wallet, error) {
	// Check if wallet already exists
	if existing, _ := s.walletRepo.GetByUserID(userID); existing != nil {
		return existing, nil
	}

	wallet := &domain.Wallet{
		ID:             uuid.New().String(),
		UserID:         userID,
		UserType:       userType,
		Balance:        0.0,
		PendingBalance: 0.0,
		Currency:       "USD",
		Status:         domain.WalletStatusActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.walletRepo.Create(wallet); err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return wallet, nil
}

func (s *paymentService) GetWallet(userID string) (*domain.Wallet, error) {
	return s.walletRepo.GetByUserID(userID)
}

func (s *paymentService) GetBalance(userID string) (*domain.WalletBalance, error) {
	wallet, err := s.walletRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	return &domain.WalletBalance{
		Balance:        wallet.Balance,
		PendingBalance: wallet.PendingBalance,
		Currency:       wallet.Currency,
	}, nil
}

// Payment processing
func (s *paymentService) ProcessPayment(req domain.ProcessPaymentRequest) (*domain.PaymentResponse, error) {
	// Create transaction record
	transactionID := uuid.New().String()

	// Get payer wallet
	payerWallet, err := s.walletRepo.GetByUserID(req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("payer wallet not found: %w", err)
	}

	// Get payment method to determine type
	paymentMethod, err := s.paymentMethodRepo.GetByID(req.PaymentMethodID)
	if err != nil {
		return nil, fmt.Errorf("payment method not found: %w", err)
	}

	transaction := &domain.Transaction{
		ID:              transactionID,
		FromWalletID:    &payerWallet.ID,
		Type:            domain.TxTypePayment,
		Status:          domain.TxStatusPending,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Description:     req.Description,
		OrderID:         &req.OrderID,
		PaymentMethodID: &req.PaymentMethodID,
		Metadata:        req.Metadata,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.transactionRepo.Create(transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Process payment based on method type
	var paymentResult *domain.PaymentResponse
	switch paymentMethod.Type {
	case domain.PaymentTypeCard:
		paymentResult, err = s.processCardPayment(req, transaction)
	case domain.PaymentTypeDigitalWallet:
		paymentResult, err = s.processWalletPayment(req, transaction, payerWallet)
	case domain.PaymentTypeBankAccount:
		paymentResult, err = s.processBankPayment(req, transaction)
	default:
		return nil, errors.New("unsupported payment method")
	}

	if err != nil {
		transaction.Status = domain.TxStatusFailed
		transaction.UpdatedAt = time.Now()
		s.transactionRepo.Update(transaction)
		return nil, err
	}

	// Update transaction with success
	transaction.Status = domain.TxStatusCompleted
	transaction.Fee = paymentResult.Fee
	transaction.NetAmount = paymentResult.NetAmount
	now := time.Now()
	transaction.ProcessedAt = &now
	transaction.UpdatedAt = now

	if err := s.transactionRepo.Update(transaction); err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	paymentResult.TransactionID = transactionID
	return paymentResult, nil
}

func (s *paymentService) ProcessRefund(req domain.RefundRequest) (*domain.PaymentResponse, error) {
	// Find original transaction
	originalTx, err := s.transactionRepo.GetByID(req.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("original transaction not found: %w", err)
	}

	// Create refund transaction
	refundTransaction := &domain.Transaction{
		ID:          uuid.New().String(),
		ToWalletID:  originalTx.FromWalletID, // Refund goes back to original payer
		Type:        domain.TxTypeRefund,
		Status:      domain.TxStatusPending,
		Amount:      req.Amount,
		Currency:    originalTx.Currency,
		Description: fmt.Sprintf("Refund: %s", req.Reason),
		Reference:   req.TransactionID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.transactionRepo.Create(refundTransaction); err != nil {
		return nil, fmt.Errorf("failed to create refund transaction: %w", err)
	}

	// Update wallet balance if refunding to wallet
	if originalTx.FromWalletID != nil {
		wallet, err := s.walletRepo.GetByID(*originalTx.FromWalletID)
		if err != nil {
			return nil, err
		}

		wallet.Balance += req.Amount
		wallet.UpdatedAt = time.Now()

		if err := s.walletRepo.Update(wallet); err != nil {
			return nil, err
		}
	}

	// Complete refund transaction
	refundTransaction.Status = domain.TxStatusCompleted
	now := time.Now()
	refundTransaction.ProcessedAt = &now
	refundTransaction.UpdatedAt = now
	s.transactionRepo.Update(refundTransaction)

	return &domain.PaymentResponse{
		TransactionID: refundTransaction.ID,
		Status:        domain.TxStatusCompleted,
		Amount:        req.Amount,
		NetAmount:     req.Amount,
		ProcessedAt:   refundTransaction.ProcessedAt,
	}, nil
}

func (s *paymentService) ProcessTransfer(req domain.TransferRequest) (*domain.PaymentResponse, error) {
	// Get sender wallet
	senderWallet, err := s.walletRepo.GetByUserID(req.FromUserID)
	if err != nil {
		return nil, fmt.Errorf("sender wallet not found: %w", err)
	}

	// Check sufficient balance
	if senderWallet.Balance < req.Amount {
		return nil, errors.New("insufficient balance")
	}

	// Get receiver wallet
	receiverWallet, err := s.walletRepo.GetByUserID(req.ToUserID)
	if err != nil {
		return nil, fmt.Errorf("receiver wallet not found: %w", err)
	}

	transactionID := uuid.New().String()

	// Debit sender
	senderWallet.Balance -= req.Amount
	senderWallet.UpdatedAt = time.Now()

	// Credit receiver
	receiverWallet.Balance += req.Amount
	receiverWallet.UpdatedAt = time.Now()

	// Create transfer transaction
	transaction := &domain.Transaction{
		ID:           transactionID,
		FromWalletID: &senderWallet.ID,
		ToWalletID:   &receiverWallet.ID,
		Type:         domain.TxTypeTransfer,
		Status:       domain.TxStatusCompleted,
		Amount:       req.Amount,
		Currency:     senderWallet.Currency,
		Description:  req.Description,
		Metadata:     req.Metadata,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	now := time.Now()
	transaction.ProcessedAt = &now

	// Save all changes
	if err := s.walletRepo.Update(senderWallet); err != nil {
		return nil, err
	}
	if err := s.walletRepo.Update(receiverWallet); err != nil {
		return nil, err
	}
	if err := s.transactionRepo.Create(transaction); err != nil {
		return nil, err
	}

	return &domain.PaymentResponse{
		TransactionID: transactionID,
		Status:        domain.TxStatusCompleted,
		Amount:        req.Amount,
		NetAmount:     req.Amount,
		ProcessedAt:   transaction.ProcessedAt,
	}, nil
}

func (s *paymentService) ProcessTopUp(req domain.TopUpRequest) (*domain.PaymentResponse, error) {
	// Get wallet
	wallet, err := s.walletRepo.GetByUserID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	transactionID := uuid.New().String()

	// Create transaction
	transaction := &domain.Transaction{
		ID:              transactionID,
		ToWalletID:      &wallet.ID,
		Type:            domain.TxTypeTopUp,
		Status:          domain.TxStatusPending,
		Amount:          req.Amount,
		Currency:        wallet.Currency,
		Description:     "Wallet top-up",
		PaymentMethodID: &req.PaymentMethodID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.transactionRepo.Create(transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Process payment via external provider
	// In production, use actual payment gateway
	// For now, simulate successful payment

	// Update wallet balance
	wallet.Balance += req.Amount
	wallet.UpdatedAt = time.Now()

	if err := s.walletRepo.Update(wallet); err != nil {
		return nil, err
	}

	// Complete transaction
	transaction.Status = domain.TxStatusCompleted
	now := time.Now()
	transaction.ProcessedAt = &now
	transaction.UpdatedAt = now
	s.transactionRepo.Update(transaction)

	return &domain.PaymentResponse{
		TransactionID: transactionID,
		Status:        domain.TxStatusCompleted,
		Amount:        req.Amount,
		NetAmount:     req.Amount,
		ProcessedAt:   transaction.ProcessedAt,
	}, nil
}

func (s *paymentService) ProcessWithdrawal(req domain.WithdrawalRequest) (*domain.PaymentResponse, error) {
	// Get wallet
	wallet, err := s.walletRepo.GetByUserID(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	// Check sufficient balance
	if wallet.Balance < req.Amount {
		return nil, errors.New("insufficient balance")
	}

	transactionID := uuid.New().String()

	// Create transaction
	transaction := &domain.Transaction{
		ID:              transactionID,
		FromWalletID:    &wallet.ID,
		Type:            domain.TxTypeWithdrawal,
		Status:          domain.TxStatusPending,
		Amount:          req.Amount,
		Currency:        wallet.Currency,
		Description:     "Wallet withdrawal",
		PaymentMethodID: &req.PaymentMethodID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.transactionRepo.Create(transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Process withdrawal via bank transfer
	// In production, use actual bank service
	// For now, simulate successful withdrawal

	// Update wallet balance
	wallet.Balance -= req.Amount
	wallet.UpdatedAt = time.Now()

	if err := s.walletRepo.Update(wallet); err != nil {
		return nil, err
	}

	// Complete transaction
	transaction.Status = domain.TxStatusCompleted
	now := time.Now()
	transaction.ProcessedAt = &now
	transaction.UpdatedAt = now
	s.transactionRepo.Update(transaction)

	return &domain.PaymentResponse{
		TransactionID: transactionID,
		Status:        domain.TxStatusCompleted,
		Amount:        req.Amount,
		NetAmount:     req.Amount,
		ProcessedAt:   transaction.ProcessedAt,
	}, nil
}

// Payment methods
func (s *paymentService) AddPaymentMethod(userID string, req domain.AddPaymentMethodRequest) (*domain.PaymentMethod, error) {
	// Validate payment method data
	if req.Type == domain.PaymentTypeCard && req.CardToken == "" {
		return nil, errors.New("card token required")
	}
	if req.Type == domain.PaymentTypeBankAccount && req.BankInfo == nil {
		return nil, errors.New("bank information required")
	}

	paymentMethod := &domain.PaymentMethod{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      req.Type,
		Provider:  req.Provider,
		IsDefault: req.IsDefault,
		Status:    domain.PaymentStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.BankInfo != nil {
		paymentMethod.BankInfo = req.BankInfo
	}
	if req.DigitalWallet != nil {
		paymentMethod.DigitalWallet = req.DigitalWallet
	}

	if err := s.paymentMethodRepo.Create(paymentMethod); err != nil {
		return nil, fmt.Errorf("failed to add payment method: %w", err)
	}

	// Set as default if requested
	if req.IsDefault {
		s.paymentMethodRepo.SetAsDefault(userID, paymentMethod.ID)
	}

	return paymentMethod, nil
}

func (s *paymentService) GetPaymentMethods(userID string) ([]domain.PaymentMethod, error) {
	return s.paymentMethodRepo.GetByUserID(userID)
}

func (s *paymentService) SetDefaultPaymentMethod(userID, methodID string) error {
	return s.paymentMethodRepo.SetAsDefault(userID, methodID)
}

func (s *paymentService) RemovePaymentMethod(userID, methodID string) error {
	// Verify ownership
	method, err := s.paymentMethodRepo.GetByID(methodID)
	if err != nil {
		return err
	}

	if method.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.paymentMethodRepo.Delete(methodID)
}

// Transactions
func (s *paymentService) GetTransaction(transactionID string) (*domain.Transaction, error) {
	return s.transactionRepo.GetByID(transactionID)
}

func (s *paymentService) GetTransactionHistory(userID string, limit, offset int) ([]domain.Transaction, error) {
	return s.transactionRepo.GetByUserID(userID, limit, offset)
}

func (s *paymentService) GetTransactionReport(userID string, startDate, endDate time.Time) (*domain.TransactionReport, error) {
	return s.transactionRepo.GetTransactionReport(userID, startDate, endDate)
}

// Commission management
func (s *paymentService) CalculateCommission(orderID string, orderAmount, deliveryFee float64, merchantID, driverID string) (*domain.Commission, error) {
	platformFee := orderAmount * 0.03 // 3% platform fee
	merchantFee := orderAmount * 0.02 // 2% merchant fee
	driverFee := deliveryFee * 0.10   // 10% driver fee

	commission := &domain.Commission{
		ID:            uuid.New().String(),
		OrderID:       orderID,
		MerchantID:    merchantID,
		DriverID:      &driverID,
		OrderAmount:   orderAmount,
		DeliveryFee:   deliveryFee,
		PlatformFee:   platformFee,
		MerchantFee:   merchantFee,
		DriverFee:     driverFee,
		NetToMerchant: orderAmount - platformFee - merchantFee,
		NetToDriver:   deliveryFee - driverFee,
		Status:        domain.CommissionStatusPending,
		CreatedAt:     time.Now(),
	}

	if err := s.commissionRepo.Create(commission); err != nil {
		return nil, fmt.Errorf("failed to create commission: %w", err)
	}

	return commission, nil
}

func (s *paymentService) ProcessCommission(commissionID string) error {
	commission, err := s.commissionRepo.GetByID(commissionID)
	if err != nil {
		return err
	}

	commission.Status = domain.CommissionStatusProcessed
	now := time.Now()
	commission.ProcessedAt = &now

	return s.commissionRepo.Update(commission)
}

func (s *paymentService) GetMerchantCommissions(merchantID string, limit, offset int) ([]domain.Commission, error) {
	return s.commissionRepo.GetByMerchantID(merchantID, limit, offset)
}

func (s *paymentService) GetDriverCommissions(driverID string, limit, offset int) ([]domain.Commission, error) {
	return s.commissionRepo.GetByDriverID(driverID, limit, offset)
}

// Payouts
func (s *paymentService) ProcessMerchantPayout(merchantID string, amount float64) (*domain.PaymentResponse, error) {
	// Get merchant wallet
	wallet, err := s.walletRepo.GetByUserID(merchantID)
	if err != nil {
		return nil, fmt.Errorf("merchant wallet not found: %w", err)
	}

	transactionID := uuid.New().String()

	// Create payout transaction
	transaction := &domain.Transaction{
		ID:          transactionID,
		ToWalletID:  &wallet.ID,
		Type:        domain.TxTypePayout,
		Status:      domain.TxStatusCompleted,
		Amount:      amount,
		Currency:    wallet.Currency,
		Description: "Merchant payout",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	now := time.Now()
	transaction.ProcessedAt = &now

	// Update wallet balance
	wallet.Balance += amount
	wallet.UpdatedAt = time.Now()

	if err := s.walletRepo.Update(wallet); err != nil {
		return nil, err
	}

	if err := s.transactionRepo.Create(transaction); err != nil {
		return nil, err
	}

	return &domain.PaymentResponse{
		TransactionID: transactionID,
		Status:        domain.TxStatusCompleted,
		Amount:        amount,
		NetAmount:     amount,
		ProcessedAt:   transaction.ProcessedAt,
	}, nil
}

func (s *paymentService) ProcessDriverPayout(driverID string, amount float64) (*domain.PaymentResponse, error) {
	// Get driver wallet
	wallet, err := s.walletRepo.GetByUserID(driverID)
	if err != nil {
		return nil, fmt.Errorf("driver wallet not found: %w", err)
	}

	transactionID := uuid.New().String()

	// Create payout transaction
	transaction := &domain.Transaction{
		ID:          transactionID,
		ToWalletID:  &wallet.ID,
		Type:        domain.TxTypePayout,
		Status:      domain.TxStatusCompleted,
		Amount:      amount,
		Currency:    wallet.Currency,
		Description: "Driver payout",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	now := time.Now()
	transaction.ProcessedAt = &now

	// Update wallet balance
	wallet.Balance += amount
	wallet.UpdatedAt = time.Now()

	if err := s.walletRepo.Update(wallet); err != nil {
		return nil, err
	}

	if err := s.transactionRepo.Create(transaction); err != nil {
		return nil, err
	}

	return &domain.PaymentResponse{
		TransactionID: transactionID,
		Status:        domain.TxStatusCompleted,
		Amount:        amount,
		NetAmount:     amount,
		ProcessedAt:   transaction.ProcessedAt,
	}, nil
}

// Helper methods for payment processing
func (s *paymentService) processCardPayment(req domain.ProcessPaymentRequest, transaction *domain.Transaction) (*domain.PaymentResponse, error) {
	// In production, integrate with Stripe or other payment processor
	// For now, simulate successful payment

	fee := req.Amount * 0.029 // 2.9% typical card fee
	netAmount := req.Amount - fee

	return &domain.PaymentResponse{
		Status:    domain.TxStatusCompleted,
		Amount:    req.Amount,
		Fee:       fee,
		NetAmount: netAmount,
	}, nil
}

func (s *paymentService) processWalletPayment(req domain.ProcessPaymentRequest, transaction *domain.Transaction, wallet *domain.Wallet) (*domain.PaymentResponse, error) {
	// Check sufficient balance
	if wallet.Balance < req.Amount {
		return nil, errors.New("insufficient wallet balance")
	}

	// Deduct from wallet
	wallet.Balance -= req.Amount
	wallet.UpdatedAt = time.Now()

	if err := s.walletRepo.Update(wallet); err != nil {
		return nil, err
	}

	return &domain.PaymentResponse{
		Status:    domain.TxStatusCompleted,
		Amount:    req.Amount,
		Fee:       0.0,
		NetAmount: req.Amount,
	}, nil
}

func (s *paymentService) processBankPayment(req domain.ProcessPaymentRequest, transaction *domain.Transaction) (*domain.PaymentResponse, error) {
	// In production, integrate with ACH/bank transfer service
	// For now, simulate successful payment

	fee := 0.50 // Flat ACH fee
	netAmount := req.Amount - fee

	return &domain.PaymentResponse{
		Status:    domain.TxStatusCompleted,
		Amount:    req.Amount,
		Fee:       fee,
		NetAmount: netAmount,
	}, nil
}

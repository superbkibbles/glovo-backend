package client

import (
	"glovo-backend/services/payment-service/internal/domain"
)

// Mock Stripe Service
type mockStripeService struct{}

func NewMockStripeService() domain.StripeService {
	return &mockStripeService{}
}

func (m *mockStripeService) ProcessCardPayment(token string, amount float64, currency string) (*domain.StripePaymentResult, error) {
	return &domain.StripePaymentResult{
		ChargeID:  "ch_mock_charge_123",
		Status:    "succeeded",
		Amount:    int64(amount * 100), // Convert to cents
		Currency:  currency,
		Reference: "mock_payment_ref",
	}, nil
}

func (m *mockStripeService) CreateCustomer(userID, email string) (*domain.StripeCustomer, error) {
	return &domain.StripeCustomer{
		ID:    "cus_mock_customer_" + userID,
		Email: email,
	}, nil
}

func (m *mockStripeService) RefundPayment(chargeID string, amount float64) (*domain.StripeRefund, error) {
	return &domain.StripeRefund{
		ID:       "re_mock_refund_123",
		Amount:   int64(amount * 100), // Convert to cents
		Status:   "succeeded",
		ChargeID: chargeID,
	}, nil
}

// Mock Bank Service
type mockBankService struct{}

func NewMockBankService() domain.BankService {
	return &mockBankService{}
}

func (m *mockBankService) ProcessACHTransfer(accountInfo domain.BankAccountInfo, amount float64) (*domain.BankTransferResult, error) {
	return &domain.BankTransferResult{
		TransferID: "ach_mock_transfer_123",
		Status:     "completed",
		Amount:     amount,
		Reference:  "mock_bank_ref",
	}, nil
}

func (m *mockBankService) ValidateBankAccount(accountInfo domain.BankAccountInfo) error {
	// Mock validation - always succeeds
	return nil
}

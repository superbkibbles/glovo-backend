package application

import (
	"context"
)

// OTPRepository defines OTP storage and delivery
type OTPRepository interface {
	StorePhoneOTP(ctx context.Context, phone, code string) error
	StoreEmailOTP(ctx context.Context, email, code string) error
	VerifyAndConsumePhoneOTP(ctx context.Context, phone, code string) (bool, error)
	VerifyAndConsumeEmailOTP(ctx context.Context, email, code string) (bool, error)
	SendPhoneOTP(ctx context.Context, phone, code string) error
	SendEmailOTP(ctx context.Context, email, code string) error
}

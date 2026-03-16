package repository

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/mendmzury/food-delivery/pkg/database"
	"github.com/mendmzury/food-delivery/pkg/logger"
	"github.com/redis/go-redis/v9"
)

const (
	otpKeyPrefix   = "otp:"
	otpPhonePrefix = "otp:phone:"
	otpEmailPrefix = "otp:email:"
	otpTTL         = 5 * time.Minute
)

type OTPRepository struct {
	redis *database.RedisDB
}

func NewOTPRepository(redis *database.RedisDB) *OTPRepository {
	return &OTPRepository{redis: redis}
}

func (r *OTPRepository) StorePhoneOTP(ctx context.Context, phone, code string) error {
	key := otpPhonePrefix + phone
	return r.redis.Set(ctx, key, code, otpTTL)
}

func (r *OTPRepository) StoreEmailOTP(ctx context.Context, email, code string) error {
	key := otpEmailPrefix + email
	return r.redis.Set(ctx, key, code, otpTTL)
}

func (r *OTPRepository) GetPhoneOTP(ctx context.Context, phone string) (string, error) {
	key := otpPhonePrefix + phone
	code, err := r.redis.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	return code, nil
}

func (r *OTPRepository) GetEmailOTP(ctx context.Context, email string) (string, error) {
	key := otpEmailPrefix + email
	code, err := r.redis.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}
	return code, nil
}

func (r *OTPRepository) VerifyPhoneOTP(ctx context.Context, phone, code string) (bool, error) {
	stored, err := r.GetPhoneOTP(ctx, phone)
	if err != nil {
		return false, err
	}
	if stored == "" {
		return false, nil
	}
	valid := stored == code
	if valid {
		key := otpPhonePrefix + phone
		_ = r.redis.Delete(ctx, key)
	}
	return valid, nil
}

func (r *OTPRepository) VerifyEmailOTP(ctx context.Context, email, code string) (bool, error) {
	stored, err := r.GetEmailOTP(ctx, email)
	if err != nil {
		return false, err
	}
	if stored == "" {
		return false, nil
	}
	valid := stored == code
	if valid {
		key := otpEmailPrefix + email
		_ = r.redis.Delete(ctx, key)
	}
	return valid, nil
}

func (r *OTPRepository) VerifyAndConsumePhoneOTP(ctx context.Context, phone, code string) (bool, error) {
	return r.VerifyPhoneOTP(ctx, phone, code)
}

func (r *OTPRepository) VerifyAndConsumeEmailOTP(ctx context.Context, email, code string) (bool, error) {
	return r.VerifyEmailOTP(ctx, email, code)
}

func (r *OTPRepository) SendPhoneOTP(ctx context.Context, phone, code string) error {
	// In dev, log the OTP. In production, integrate Twilio or similar.
	logger.Info().
		Str("phone", phone).
		Str("otp", code).
		Msg("OTP for phone (configure TWILIO_* for production SMS)")
	// TODO: Integrate Twilio when TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, TWILIO_PHONE_NUMBER are set
	return nil
}

func (r *OTPRepository) SendEmailOTP(ctx context.Context, email, code string) error {
	// In dev, log the OTP. In production, integrate SendGrid/SMTP.
	logger.Info().
		Str("email", email).
		Str("otp", code).
		Msg("OTP for email (configure SMTP_* or SENDGRID_* for production)")
	// TODO: Integrate SendGrid/SMTP when configured
	return nil
}

func GenerateOTPCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}

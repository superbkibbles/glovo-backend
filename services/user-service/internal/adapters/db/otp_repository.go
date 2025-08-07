package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"glovo-backend/services/user-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

type otpRepository struct {
	redis *redis.Client
}

func NewOTPRepository(redisClient *redis.Client) domain.OTPRepository {
	return &otpRepository{redis: redisClient}
}

func (r *otpRepository) Store(otp *domain.OTP) error {
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", otp.PhoneNumber)

	// Serialize OTP to JSON
	data, err := json.Marshal(otp)
	if err != nil {
		return err
	}

	// Store with expiration time
	ttl := time.Until(otp.ExpiresAt)
	return r.redis.Set(ctx, key, data, ttl).Err()
}

func (r *otpRepository) GetByPhoneNumber(phoneNumber string) (*domain.OTP, error) {
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", phoneNumber)

	data, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var otp domain.OTP
	if err := json.Unmarshal([]byte(data), &otp); err != nil {
		return nil, err
	}

	return &otp, nil
}

func (r *otpRepository) Delete(phoneNumber string) error {
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", phoneNumber)
	return r.redis.Del(ctx, key).Err()
}

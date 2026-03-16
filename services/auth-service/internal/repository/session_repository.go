package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mendmzury/food-delivery/pkg/database"
	"github.com/mendmzury/food-delivery/services/auth-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

const sessionKeyPrefix = "session:"

type SessionRepository struct {
	redis *database.RedisDB
}

func NewSessionRepository(redis *database.RedisDB) *SessionRepository {
	return &SessionRepository{redis: redis}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session, expiry time.Duration) error {
	key := sessionKeyPrefix + session.Token
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	return r.redis.Set(ctx, key, data, expiry)
}

func (r *SessionRepository) Get(ctx context.Context, token string) (*domain.Session, error) {
	key := sessionKeyPrefix + token
	data, err := r.redis.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var session domain.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	return &session, nil
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	key := sessionKeyPrefix + token
	return r.redis.Delete(ctx, key)
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	// Placeholder - in production track user sessions
	return nil
}

func (r *SessionRepository) Exists(ctx context.Context, token string) (bool, error) {
	key := sessionKeyPrefix + token
	return r.redis.Exists(ctx, key)
}

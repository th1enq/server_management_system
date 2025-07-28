package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/th1enq/server_management_system/internal/configs"
	"go.uber.org/zap"
)

var (
	ErrCacheMiss = errors.New("cache miss")
)

type CacheClient interface {
	Set(ctx context.Context, key string, data any, ttl time.Duration) error
	Get(ctx context.Context, key string, dest any) error
	Del(ctx context.Context, key string) error
	SADD(ctx context.Context, key string, members ...string) error
	SMEMBERS(ctx context.Context, key string) ([]string, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	HMGet(ctx context.Context, key string) (map[string]string, error)
	HSET(ctx context.Context, key string, values map[string]string) error
}

type redisClient struct {
	client *redis.Client
	logger *zap.Logger
}

func (r *redisClient) Set(ctx context.Context, key string, data any, ttl time.Duration) error {
	byte, err := json.Marshal(data)
	if err != nil {
		r.logger.Error("Failed to marshal data for Redis", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to marshal data for Redis: %w", err)
	}
	if err := r.client.Set(ctx, key, byte, ttl).Err(); err != nil {
		r.logger.Error("Failed to set data in Redis", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to set data in Redis: %w", err)
	}
	r.logger.Info("Data set in Redis successfully", zap.String("key", key), zap.Duration("ttl", ttl))
	return nil
}

func (r *redisClient) Get(ctx context.Context, key string, dest any) error {
	byte, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			r.logger.Warn("Cache miss for key", zap.String("key", key))
			return ErrCacheMiss
		}
		r.logger.Error("Failed to get data from Redis", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to get data from Redis: %w", err)
	}

	if err := json.Unmarshal(byte, dest); err != nil {
		r.logger.Error("Failed to unmarshal data from Redis", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to unmarshal data from Redis: %w", err)
	}
	return nil
}

func (r *redisClient) Del(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		r.logger.Error("Failed to delete data from Redis", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to delete data from Redis: %w", err)
	}
	return nil
}

func (r *redisClient) SADD(ctx context.Context, key string, members ...string) error {
	if err := r.client.SAdd(ctx, key, members).Err(); err != nil {
		r.logger.Error("Failed to add members to Redis set", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to add members to Redis set: %w", err)
	}
	return nil
}

func (r *redisClient) SMEMBERS(ctx context.Context, key string) ([]string, error) {
	members, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		r.logger.Error("Failed to get members from Redis set", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("failed to get members from Redis set: %w", err)
	}
	return members, nil
}

func (r *redisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
		r.logger.Error("Failed to set expiration for Redis key", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to set expiration for Redis key: %w", err)
	}
	return nil
}

func (r *redisClient) HMGet(ctx context.Context, key string) (map[string]string, error) {
	values, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			r.logger.Warn("Cache miss for hash key", zap.String("key", key))
			return nil, ErrCacheMiss
		}
		r.logger.Error("Failed to get hash fields from Redis", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("failed to get hash fields from Redis: %w", err)
	}
	return values, nil
}

func (r *redisClient) HSET(ctx context.Context, key string, values map[string]string) error {
	if err := r.client.HSet(ctx, key, values).Err(); err != nil {
		r.logger.Error("Failed to set hash fields in Redis", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to set hash fields in Redis: %w", err)
	}
	return nil
}

func NewCache(cfg configs.Cache, logger *zap.Logger) (CacheClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Int("db", cfg.DB),
	)
	return &redisClient{
		client: rdb,
		logger: logger,
	}, nil
}

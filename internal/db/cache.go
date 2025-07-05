package db

import (
	"context"
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

type IRedisClient interface {
	Set(ctx context.Context, key string, data any, ttl time.Duration) error
	Get(ctx context.Context, key string) (any, error)
	AddToSet(ctx context.Context, key string, data ...any) error
	IsDataInSet(ctx context.Context, key string, data any) (bool, error)
	RemoveFromSet(ctx context.Context, key string, data ...any) error
}

type redisClient struct {
	client *redis.Client
	logger *zap.Logger
}

func (r *redisClient) Set(ctx context.Context, key string, data any, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		r.logger.Error("Failed to set data in Redis", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to set data in Redis: %w", err)
	}
	return nil
}

func (r *redisClient) Get(ctx context.Context, key string) (any, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		r.logger.Error("Failed to get data from Redis", zap.String("key", key), zap.Error(err))
		return nil, fmt.Errorf("failed to get data from Redis: %w", err)
	}
	return data, nil
}

func (r *redisClient) AddToSet(ctx context.Context, key string, data ...any) error {
	if err := r.client.SAdd(ctx, key, data...).Err(); err != nil {
		r.logger.Error("Failed to add data to Redis set", zap.String("key", key), zap.Any("data", data), zap.Error(err))
		return fmt.Errorf("failed to add data to Redis set: %w", err)
	}
	return nil
}

func (r *redisClient) IsDataInSet(ctx context.Context, key string, data any) (bool, error) {
	exists, err := r.client.SIsMember(ctx, key, data).Result()
	if err != nil {
		r.logger.Error("Failed to check if data is in Redis set", zap.String("key", key), zap.Any("data", data), zap.Error(err))
		return false, fmt.Errorf("failed to check if data is in Redis set: %w", err)
	}
	return exists, nil
}

func (r *redisClient) RemoveFromSet(ctx context.Context, key string, data ...any) error {
	if err := r.client.SRem(ctx, key, data...).Err(); err != nil {
		r.logger.Error("Failed to remove data from Redis set", zap.String("key", key), zap.Any("data", data), zap.Error(err))
		return fmt.Errorf("failed to remove data from Redis set: %w", err)
	}
	return nil
}

func NewCache(cfg configs.Cache, logger *zap.Logger) (IRedisClient, error) {
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

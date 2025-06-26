package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/th1enq/server_management_system/internal/configs"
	"go.uber.org/zap"
)

func LoadCache(cfg configs.Cache, logger *zap.Logger) (*redis.Client, func(), error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	cleanup := func() {
		redisClient.Close()
	}

	logger.Info("Redis connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Int("db", cfg.DB),
	)
	return redisClient, cleanup, nil
}

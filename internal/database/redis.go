package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/th1enq/server_management_system/internal/config"
	"github.com/th1enq/server_management_system/pkg/logger"
	"go.uber.org/zap"
)

var RedisClient *redis.Client

func LoadRedis(config *config.Config) error {
	cfg := config.Redis
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Int("db", cfg.DB),
	)
	return nil
}

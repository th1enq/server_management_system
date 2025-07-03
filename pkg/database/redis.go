package database

import (
	"context"
	"fmt"
	"server_management_system/pkg/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func LoadRedis(cfg config.Redis, log *zap.Logger) (*redis.Client, func(), error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Error("Failed to connect to Redis", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	cleanup := func() {
		log.Info("Closing Redis connection")
		redisClient.Close()
	}

	log.Info("Redis connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Int("db", cfg.DB),
	)
	return redisClient, cleanup, nil
}

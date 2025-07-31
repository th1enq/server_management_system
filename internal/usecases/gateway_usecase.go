package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/dto"
	"github.com/th1enq/server_management_system/internal/infrastructure/cache"
	"github.com/th1enq/server_management_system/internal/infrastructure/mq/producer"
	"go.uber.org/zap"
)

type GatewayUseCase interface {
	ProcessServerMetrics(ctx context.Context, metrics dto.MetricsRequest) error
}

type gatewayUseCase struct {
	logger               *zap.Logger
	serverUseCase        ServerUseCase
	redisCache           cache.CacheClient
	monitoringProducer   producer.MonitoringMessageProducer
	statusChangeProducer producer.StatusChangeMessageProducer
}

func NewGatewayUseCase(
	serverUseCase ServerUseCase,
	redisCache cache.CacheClient,
	monitoringProducer producer.MonitoringMessageProducer,
	statusChangeProducer producer.StatusChangeMessageProducer,
	logger *zap.Logger,
) GatewayUseCase {
	return &gatewayUseCase{
		serverUseCase:        serverUseCase,
		redisCache:           redisCache,
		monitoringProducer:   monitoringProducer,
		statusChangeProducer: statusChangeProducer,
		logger:               logger,
	}
}

func (s *gatewayUseCase) ProcessServerMetrics(ctx context.Context, metrics dto.MetricsRequest) error {
	s.logger.Info("Processing server metrics in Gateway",
		zap.String("server_id", metrics.ServerID),
	)

	cacheKey := fmt.Sprintf("heartbeat:%s", metrics.ServerID)
	var intervalCheckTime int64
	if err := s.redisCache.Get(ctx, cacheKey, &intervalCheckTime); err == nil {
		expireTime := time.Duration(1.5*float64(intervalCheckTime)) * time.Second
		s.redisCache.Expire(ctx, cacheKey, expireTime)

		if err := s.monitoringProducer.Produce(ctx, producer.Message{
			ServerID:  metrics.ServerID,
			OldStatus: entity.ServerStatusOn,
			NewStatus: entity.ServerStatusOn,
			Timestamp: metrics.Timestamp,
		}); err != nil {
			s.logger.Error("Failed to produce monitoring message", zap.Error(err))
			return fmt.Errorf("failed to produce monitoring message: %w", err)
		}

		return nil
	}
	server, err := s.serverUseCase.GetServerByID(ctx, metrics.ServerID)
	if err != nil {
		s.logger.Error("Failed to get server by ID", zap.String("server_id", metrics.ServerID), zap.Error(err))
		return fmt.Errorf("server not found")
	}

	intervalCheckTime = server.IntervalTime
	expireTime := time.Duration(1.5*float64(intervalCheckTime)) * time.Second

	s.redisCache.Set(ctx, cacheKey, intervalCheckTime, expireTime)

	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	metrics.Timestamp = metrics.Timestamp.In(loc)

	if err := s.statusChangeProducer.Produce(ctx, producer.StatusChangeMessage{
		ServerID:  metrics.ServerID,
		OldStatus: server.Status,
		NewStatus: entity.ServerStatusOn,
		Timestamp: metrics.Timestamp,
		Interval:  intervalCheckTime,
	}); err != nil {
		s.logger.Error("Failed to produce status change message", zap.Error(err))
		return fmt.Errorf("failed to produce status change message: %w", err)
	}
	if err := s.monitoringProducer.Produce(ctx, producer.Message{
		ServerID:  metrics.ServerID,
		OldStatus: server.Status,
		NewStatus: entity.ServerStatusOn,
		Timestamp: metrics.Timestamp,
	}); err != nil {
		s.logger.Error("Failed to produce monitoring message", zap.Error(err))
		return fmt.Errorf("failed to produce monitoring message: %w", err)
	}
	return nil
}

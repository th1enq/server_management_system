package consumers

import (
	"context"

	"github.com/th1enq/server_management_system/internal/infrastructure/mq/producer"
	"github.com/th1enq/server_management_system/internal/usecases"
	"github.com/th1enq/server_management_system/internal/utils"
	"go.uber.org/zap"
)

type MetricsUpdate interface {
	Handle(ctx context.Context, event producer.MonitoringMessage) error
}

type metricsUpdate struct {
	metricsUseCase usecases.MetricsUseCase
	logger         *zap.Logger
}

func NewMetricsUpdate(
	metricsUseCase usecases.MetricsUseCase,
	logger *zap.Logger,
) MetricsUpdate {
	return &metricsUpdate{
		metricsUseCase: metricsUseCase,
		logger:         logger,
	}
}

func (m metricsUpdate) Handle(
	ctx context.Context,
	event producer.MonitoringMessage,
) error {
	logger := utils.LoggerWithContext(ctx, m.logger).With(zap.Any("event", event))
	logger.Info("Received metrics task")

	if err := m.metricsUseCase.Insert(ctx, event); err != nil {
		logger.Error("Failed to handle metrics update", zap.Error(err))
		return err
	}

	return nil
}

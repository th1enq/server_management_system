package consumers

import (
	"context"

	"github.com/th1enq/server_management_system/internal/infrastructure/mq/producer"
	"github.com/th1enq/server_management_system/internal/usecases"
	"github.com/th1enq/server_management_system/internal/utils"
	"go.uber.org/zap"
)

type uptimeCalculatorConsumer interface {
	Handle(ctx context.Context, event producer.StatusChangeMessage) error
}

type uptimeCalculator struct {
	logger             *zap.Logger
	healthCheckUseCase usecases.HealthCheckUseCase
}

func NewUptimeCalculator(
	logger *zap.Logger,
	healthCheckUseCase usecases.HealthCheckUseCase,
) uptimeCalculatorConsumer {
	return &uptimeCalculator{
		logger:             logger,
		healthCheckUseCase: healthCheckUseCase,
	}
}

func (s *uptimeCalculator) Handle(
	ctx context.Context,
	event producer.StatusChangeMessage,
) error {
	logger := utils.LoggerWithContext(ctx, s.logger)

	logger.Info("Handling server status change for uptime calculation",
		zap.String("server_id", event.ServerID),
		zap.String("old_status", string(event.OldStatus)),
		zap.String("new_status", string(event.NewStatus)),
		zap.Time("timestamp", event.Timestamp),
	)

	if err := s.healthCheckUseCase.InsertUptime(ctx, event); err != nil {
		logger.Error("Failed to update uptime", zap.Error(err))
		return err
	}

	logger.Info("Successfully processed status change for uptime tracking",
		zap.String("server_id", event.ServerID),
	)

	return nil
}

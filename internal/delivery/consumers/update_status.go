package consumers

import (
	"context"

	"github.com/th1enq/server_management_system/internal/infrastructure/mq/producer"
	"github.com/th1enq/server_management_system/internal/usecases"
	"github.com/th1enq/server_management_system/internal/utils"
	"go.uber.org/zap"
)

type UpdateStatus interface {
	Handle(ctx context.Context, event producer.MonitoringMessage) error
}

type updateStatus struct {
	serverUseCase usecases.ServerUseCase
	logger        *zap.Logger
}

func NewUpdateStatus(
	serverUseCase usecases.ServerUseCase,
	logger *zap.Logger,
) UpdateStatus {
	return &updateStatus{
		serverUseCase: serverUseCase,
		logger:        logger,
	}
}

func (m updateStatus) Handle(
	ctx context.Context,
	event producer.MonitoringMessage,
) error {
	logger := utils.LoggerWithContext(ctx, m.logger)
	logger.Info("Handling server status off")

	if err := m.serverUseCase.ConsumeMessage(ctx, event); err != nil {
		logger.Error("Failed to handle server status off", zap.Error(err))
		return err
	}

	return nil
}

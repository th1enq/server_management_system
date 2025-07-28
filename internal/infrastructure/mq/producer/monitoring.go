package producer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/th1enq/server_management_system/internal/utils"
	"go.uber.org/zap"
)

const (
	MessageMonitoring = "monitoring_event"
)

type MonitoringMessage struct {
	ServerID string `json:"server_id" binding:"required"`
	CPU      int    `json:"cpu" binding:"required,gte=0"`
	RAM      int    `json:"ram" binding:"required,gte=0"`
	Disk     int    `json:"disk" binding:"required,gte=0"`
}

type MonitoringMessageProducer interface {
	Produce(ctx context.Context, msg MonitoringMessage) error
}

type monitoringMessageProducer struct {
	client Client
	logger *zap.Logger
}

func NewMonitoringMessageProducer(
	client Client,
	logger *zap.Logger,
) MonitoringMessageProducer {
	return &monitoringMessageProducer{
		client: client,
		logger: logger,
	}
}

func (m monitoringMessageProducer) Produce(
	ctx context.Context,
	msg MonitoringMessage,
) error {
	logger := utils.LoggerWithContext(ctx, m.logger)

	eventBytes, err := json.Marshal(msg)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to marshal message")
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = m.client.Produce(ctx, MessageMonitoring, eventBytes)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to produce message")
		return fmt.Errorf("failed to produce message: %w", err)
	}
	logger.Info("Produced monitoring task message", zap.Any("message", msg))
	return nil
}

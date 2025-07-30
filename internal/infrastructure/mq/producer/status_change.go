package producer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/th1enq/server_management_system/internal/utils"
	"go.uber.org/zap"
)

const (
	MessageStatusChange = "status_change_event"
)

type StatusChangeMessageProducer interface {
	Produce(ctx context.Context, msg StatusChangeMessage) error
}

type statusChangeMessageProducer struct {
	client Client
	logger *zap.Logger
}

func NewStatusChangeMessageProducer(
	client Client,
	logger *zap.Logger,
) StatusChangeMessageProducer {
	return &statusChangeMessageProducer{
		client: client,
		logger: logger,
	}
}

func (s statusChangeMessageProducer) Produce(
	ctx context.Context,
	msg StatusChangeMessage,
) error {
	logger := utils.LoggerWithContext(ctx, s.logger)

	eventBytes, err := json.Marshal(msg)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to marshal status change message")
		return fmt.Errorf("failed to marshal status change message: %w", err)
	}

	err = s.client.Produce(ctx, MessageStatusChange, eventBytes)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to produce status change message")
		return fmt.Errorf("failed to produce status change message: %w", err)
	}

	logger.Info("Produced status change message",
		zap.String("server_id", msg.ServerID),
		zap.String("old_status", string(msg.OldStatus)),
		zap.String("new_status", string(msg.NewStatus)),
	)
	return nil
}
